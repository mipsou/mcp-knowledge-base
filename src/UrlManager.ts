// UrlManager.ts
// Gestion du cycle de vie des URLs : suggest → approve/reject → fetch → markdown → index

import * as fsp from 'fs/promises';
import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';
import axios from 'axios';
import TurndownService from 'turndown';
import { KNOWLEDGE_BASES_ROOT_DIR, PENDING_URLS_FILE } from './config.js';
import { logger } from './logger.js';

export type UrlStatus = 'pending' | 'approved' | 'rejected';

export interface PendingUrl {
  id: string;
  url: string;
  knowledge_base: string;
  reason: string;
  suggested_at: string;
  status: UrlStatus;
}

function sanitizeUrlForFilename(url: string): string {
  return url
    .replace(/https?:\/\//g, '')
    .replace(/[/:?#]/g, '_')
    .replace(/_{2,}/g, '_')
    .replace(/^_|_$/g, '')
    .slice(0, 200);
}

async function loadPendingUrls(): Promise<PendingUrl[]> {
  try {
    const raw = await fsp.readFile(PENDING_URLS_FILE, 'utf-8');
    return JSON.parse(raw) as PendingUrl[];
  } catch {
    return [];
  }
}

async function savePendingUrls(urls: PendingUrl[]): Promise<void> {
  const dir = path.dirname(PENDING_URLS_FILE);
  if (!fs.existsSync(dir)) {
    await fsp.mkdir(dir, { recursive: true });
  }
  await fsp.writeFile(PENDING_URLS_FILE, JSON.stringify(urls, null, 2), 'utf-8');
}

async function fetchAndConvert(url: string): Promise<{ markdown: string; title: string }> {
  const response = await axios.get<string>(url, {
    headers: { 'User-Agent': 'knowledge-base-mcp-server/0.1.0' },
    timeout: 30_000,
    responseType: 'text',
  });
  const html: string = response.data;

  // Extraire le titre depuis la balise <title>
  const titleMatch = html.match(/<title[^>]*>([^<]*)<\/title>/i);
  const title = titleMatch ? titleMatch[1].trim() : url;

  const turndown = new TurndownService({
    headingStyle: 'atx',
    codeBlockStyle: 'fenced',
  });
  const markdown = turndown.turndown(html);
  return { markdown, title };
}

async function saveMarkdown(
  knowledgeBase: string,
  url: string,
  title: string,
  markdown: string
): Promise<string> {
  const filename = sanitizeUrlForFilename(url) + '.md';
  const dir = path.join(KNOWLEDGE_BASES_ROOT_DIR, knowledgeBase);
  if (!fs.existsSync(dir)) {
    await fsp.mkdir(dir, { recursive: true });
  }
  const filePath = path.join(dir, filename);
  const fetchedAt = new Date().toISOString();
  const frontmatter = `---\nurl: ${url}\ntitle: ${JSON.stringify(title)}\nfetched_at: ${fetchedAt}\n---\n\n`;
  await fsp.writeFile(filePath, frontmatter + markdown, 'utf-8');
  return filePath;
}

export class UrlManager {
  // Callback injecté pour éviter la dépendance circulaire avec FaissIndexManager
  private reindex: (knowledgeBase: string) => Promise<void>;

  constructor(reindex: (knowledgeBase: string) => Promise<void>) {
    this.reindex = reindex;
  }

  async suggestUrl(url: string, knowledgeBase: string, reason: string): Promise<PendingUrl> {
    const urls = await loadPendingUrls();
    const entry: PendingUrl = {
      id: crypto.randomUUID(),
      url,
      knowledge_base: knowledgeBase,
      reason,
      suggested_at: new Date().toISOString(),
      status: 'pending',
    };
    urls.push(entry);
    await savePendingUrls(urls);
    logger.info(`URL suggested: ${url} for KB "${knowledgeBase}"`);
    return entry;
  }

  async listPendingUrls(): Promise<PendingUrl[]> {
    const urls = await loadPendingUrls();
    return urls.filter((u) => u.status === 'pending');
  }

  async approveUrl(id: string): Promise<{ filePath: string; knowledgeBase: string }> {
    const urls = await loadPendingUrls();
    const entry = urls.find((u) => u.id === id);
    if (!entry) {
      throw new Error(`No pending URL found with id: ${id}`);
    }
    if (entry.status !== 'pending') {
      throw new Error(`URL ${id} has status "${entry.status}", expected "pending"`);
    }

    const { markdown, title } = await fetchAndConvert(entry.url);
    const filePath = await saveMarkdown(entry.knowledge_base, entry.url, title, markdown);

    entry.status = 'approved';
    await savePendingUrls(urls);

    await this.reindex(entry.knowledge_base);
    logger.info(`URL approved and indexed: ${entry.url} → ${filePath}`);
    return { filePath, knowledgeBase: entry.knowledge_base };
  }

  async rejectUrl(id: string): Promise<void> {
    const urls = await loadPendingUrls();
    const idx = urls.findIndex((u) => u.id === id);
    if (idx === -1) {
      throw new Error(`No pending URL found with id: ${id}`);
    }
    urls.splice(idx, 1);
    await savePendingUrls(urls);
    logger.info(`URL rejected and removed: ${id}`);
  }

  async addUrl(url: string, knowledgeBase: string): Promise<{ filePath: string }> {
    const { markdown, title } = await fetchAndConvert(url);
    const filePath = await saveMarkdown(knowledgeBase, url, title, markdown);
    await this.reindex(knowledgeBase);
    logger.info(`URL added directly and indexed: ${url} → ${filePath}`);
    return { filePath };
  }
}

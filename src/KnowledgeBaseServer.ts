// KnowledgeBaseServer.ts
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { CallToolRequestSchema, ErrorCode, ListToolsRequestSchema, McpError } from '@modelcontextprotocol/sdk/types.js';
import * as fsp from 'fs/promises';
import * as path from 'path';
import { FaissIndexManager } from './FaissIndexManager.js';
import { UrlManager } from './UrlManager.js';
import { KNOWLEDGE_BASES_ROOT_DIR } from './config.js';
import { logger } from './logger.js';

export class KnowledgeBaseServer {
  private server: Server;
  private faissManager: FaissIndexManager;
  private urlManager: UrlManager;

  constructor() {
    this.faissManager = new FaissIndexManager();
    this.urlManager = new UrlManager((kb) => this.faissManager.updateIndex(kb));
    logger.info('Initializing KnowledgeBaseServer');

    this.server = new Server(
      {
        name: 'knowledge-base-server',
        version: '0.1.0',
      },
      {
        capabilities: {
          resources: {},
          tools: {},
        },
      }
    );

    this.setupToolHandlers();

    this.server.onerror = (error) => logger.error('[MCP Error]', error);
    process.on('SIGINT', async () => {
      await this.server.close();
      process.exit(0);
    });
  }

  private setupToolHandlers() {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: [
        {
          name: 'list_knowledge_bases',
          description: 'Lists the available knowledge bases.',
          inputSchema: {
            type: 'object',
            properties: {},
            required: [],
          },
        },
        {
          name: 'retrieve_knowledge',
          description:
            'Retrieves similar chunks from the knowledge base based on a query. Optionally, if a knowledge base is specified, only that one is searched; otherwise, all available knowledge bases are considered. By default, at most 10 documents are returned with a score below a threshold of 2. A different threshold can optionally be provided.',
          inputSchema: {
            type: 'object',
            properties: {
              query: {
                type: 'string',
                description: 'The query text to use for semantic search.',
              },
              knowledge_base_name: {
                type: 'string',
                description:
                  "Optional. Name of the knowledge base to query (e.g., 'company', 'it_support', 'onboarding'). If omitted, the search is performed across all available knowledge bases.",
              },
              threshold: {
                type: 'number',
                description: 'Optional. The maximum similarity score to return.',
              },
            },
            required: ['query'],
          },
        },
        {
          name: 'suggest_url',
          description: 'Suggest a URL to add to a knowledge base. The URL is stored as pending and requires user approval before being fetched and indexed. Use this proactively when you encounter relevant official documentation.',
          inputSchema: {
            type: 'object',
            properties: {
              url: { type: 'string', description: 'The URL to suggest.' },
              knowledge_base: { type: 'string', description: 'The target knowledge base name (e.g. "infra", "podman", "stepca").' },
              reason: { type: 'string', description: 'Why this URL is relevant to the knowledge base.' },
            },
            required: ['url', 'knowledge_base', 'reason'],
          },
        },
        {
          name: 'list_pending_urls',
          description: 'List all URLs pending user approval for indexing into the knowledge base.',
          inputSchema: { type: 'object', properties: {}, required: [] },
        },
        {
          name: 'approve_url',
          description: 'Approve a pending URL: fetch its content, convert to Markdown, save to the knowledge base, and reindex. Only call this after explicit user confirmation.',
          inputSchema: {
            type: 'object',
            properties: {
              id: { type: 'string', description: 'The pending URL id returned by suggest_url.' },
            },
            required: ['id'],
          },
        },
        {
          name: 'reject_url',
          description: 'Reject and remove a pending URL without fetching it.',
          inputSchema: {
            type: 'object',
            properties: {
              id: { type: 'string', description: 'The pending URL id to reject.' },
            },
            required: ['id'],
          },
        },
        {
          name: 'add_url',
          description: 'Directly fetch a URL, convert to Markdown, save to a knowledge base, and reindex — without a pending approval step. Only use when the user explicitly provides the URL.',
          inputSchema: {
            type: 'object',
            properties: {
              url: { type: 'string', description: 'The URL to fetch and index.' },
              knowledge_base: { type: 'string', description: 'The target knowledge base name.' },
            },
            required: ['url', 'knowledge_base'],
          },
        },
      ],
    }));

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      if (request.params.name === 'list_knowledge_bases') {
        return this.handleListKnowledgeBases();
      } else if (request.params.name === 'retrieve_knowledge') {
        return this.handleRetrieveKnowledge(request.params.arguments);
      } else if (request.params.name === 'suggest_url') {
        return this.handleSuggestUrl(request.params.arguments);
      } else if (request.params.name === 'list_pending_urls') {
        return this.handleListPendingUrls();
      } else if (request.params.name === 'approve_url') {
        return this.handleApproveUrl(request.params.arguments);
      } else if (request.params.name === 'reject_url') {
        return this.handleRejectUrl(request.params.arguments);
      } else if (request.params.name === 'add_url') {
        return this.handleAddUrl(request.params.arguments);
      } else {
        throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${request.params.name}`);
      }
    });
  }

  private async handleListKnowledgeBases() {
    try {
      const entries = await fsp.readdir(KNOWLEDGE_BASES_ROOT_DIR);
      const knowledgeBases = entries.filter((entry) => !entry.startsWith('.'));
      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify(knowledgeBases, null, 2),
          },
        ],
      };
    } catch (error: any) {
      logger.error('Error listing knowledge bases:', error);
      if (error?.stack) {
        logger.error(error.stack);
      }
      return {
        content: [
          {
            type: 'text',
            text: `Error listing knowledge bases: ${error.message}`,
          },
        ],
        isError: true,
      };
    }
  }

  private async handleRetrieveKnowledge(args: any) {
    if (!args || typeof args.query !== 'string') {
      throw new McpError(ErrorCode.InvalidParams, 'Invalid arguments for retrieve_knowledge: missing query');
    }
    const query: string = args.query;
    const knowledgeBaseName: string | undefined =
      typeof args.knowledge_base_name === 'string' ? args.knowledge_base_name : undefined;
    const threshold: number | undefined =
      typeof args.threshold === 'number' ? args.threshold : undefined;

    try {
      const startTime = Date.now();
      logger.debug(`[${startTime}] handleRetrieveKnowledge started`);

      // Update FAISS index: if a specific knowledge base is provided, update only that one; otherwise update all.
      await this.faissManager.updateIndex(knowledgeBaseName);
      logger.debug(`[${Date.now()}] FAISS index update completed`);

      // Perform similarity search using the provided query.
      const similaritySearchResults = await this.faissManager.similaritySearch(query, 10, threshold);
      logger.debug(`[${Date.now()}] Similarity search completed`);

      // Build a nicely formatted markdown response including the similarity score.
      let formattedResults = '';
      if (similaritySearchResults && similaritySearchResults.length > 0) {
        formattedResults = similaritySearchResults
          .map((doc, idx) => {
            const resultHeader = `**Result ${idx + 1}:**`;
            const content = doc.pageContent.trim();
            const metadata = JSON.stringify(doc.metadata, null, 2);
            const scoreText = doc.score !== undefined ? `**Score:** ${doc.score.toFixed(2)}\n\n` : '';
            return `${resultHeader}\n\n${scoreText}${content}\n\n**Source:**\n\`\`\`json\n${metadata}\n\`\`\``;
          })
          .join('\n\n---\n\n');
      } else {
        formattedResults = '_No similar results found._';
      }
      const disclaimer = '\n\n> **Disclaimer:** The provided results might not all be relevant. Please cross-check the relevance of the information.';
      const responseText = `## Semantic Search Results\n\n${formattedResults}${disclaimer}`;

      const endTime = Date.now();
      logger.debug(`[${endTime}] handleRetrieveKnowledge completed in ${endTime - startTime}ms`);

      return {
        content: [
          {
            type: 'text',
            text: responseText,
          },
        ],
      };
    } catch (error: any) {
      logger.error('Error retrieving knowledge:', error);
      if (error?.stack) {
        logger.error(error.stack);
      }
      return {
        content: [
          {
            type: 'text',
            text: `Error retrieving knowledge: ${error.message}`,
          },
        ],
        isError: true,
      };
    }
  }

  private async handleSuggestUrl(args: any) {
    if (!args || typeof args.url !== 'string' || typeof args.knowledge_base !== 'string' || typeof args.reason !== 'string') {
      throw new McpError(ErrorCode.InvalidParams, 'suggest_url requires url, knowledge_base, and reason strings');
    }
    try {
      const entry = await this.urlManager.suggestUrl(args.url, args.knowledge_base, args.reason);
      return { content: [{ type: 'text', text: `URL suggestion enregistrée (en attente de validation) :\n\`\`\`json\n${JSON.stringify(entry, null, 2)}\n\`\`\`` }] };
    } catch (error: any) {
      logger.error('Error suggesting URL:', error);
      return { content: [{ type: 'text', text: `Error suggesting URL: ${error.message}` }], isError: true };
    }
  }

  private async handleListPendingUrls() {
    try {
      const urls = await this.urlManager.listPendingUrls();
      const text = urls.length === 0
        ? '_Aucune URL en attente._'
        : `## URLs en attente de validation (${urls.length})\n\n` + urls.map((u, i) =>
            `**${i + 1}.** \`${u.id}\`\n- URL: ${u.url}\n- KB: ${u.knowledge_base}\n- Raison: ${u.reason}\n- Proposée: ${u.suggested_at}`
          ).join('\n\n');
      return { content: [{ type: 'text', text }] };
    } catch (error: any) {
      logger.error('Error listing pending URLs:', error);
      return { content: [{ type: 'text', text: `Error listing pending URLs: ${error.message}` }], isError: true };
    }
  }

  private async handleApproveUrl(args: any) {
    if (!args || typeof args.id !== 'string') {
      throw new McpError(ErrorCode.InvalidParams, 'approve_url requires an id string');
    }
    try {
      const result = await this.urlManager.approveUrl(args.id);
      return { content: [{ type: 'text', text: `✅ URL approuvée et indexée :\n- Fichier: ${result.filePath}\n- KB: ${result.knowledgeBase}` }] };
    } catch (error: any) {
      logger.error('Error approving URL:', error);
      return { content: [{ type: 'text', text: `Error approving URL: ${error.message}` }], isError: true };
    }
  }

  private async handleRejectUrl(args: any) {
    if (!args || typeof args.id !== 'string') {
      throw new McpError(ErrorCode.InvalidParams, 'reject_url requires an id string');
    }
    try {
      await this.urlManager.rejectUrl(args.id);
      return { content: [{ type: 'text', text: `❌ URL rejetée et supprimée : ${args.id}` }] };
    } catch (error: any) {
      logger.error('Error rejecting URL:', error);
      return { content: [{ type: 'text', text: `Error rejecting URL: ${error.message}` }], isError: true };
    }
  }

  private async handleAddUrl(args: any) {
    if (!args || typeof args.url !== 'string' || typeof args.knowledge_base !== 'string') {
      throw new McpError(ErrorCode.InvalidParams, 'add_url requires url and knowledge_base strings');
    }
    try {
      const result = await this.urlManager.addUrl(args.url, args.knowledge_base);
      return { content: [{ type: 'text', text: `✅ URL ajoutée et indexée :\n- Fichier: ${result.filePath}` }] };
    } catch (error: any) {
      logger.error('Error adding URL:', error);
      return { content: [{ type: 'text', text: `Error adding URL: ${error.message}` }], isError: true };
    }
  }

  async run() {
    try {
      const transport = new StdioServerTransport();
      await this.server.connect(transport);
      logger.info('Knowledge Base MCP server running on stdio');
      await this.faissManager.initialize();
    } catch (error: any) {
      logger.error('Error during server startup:', error);
      if (error?.stack) {
        logger.error(error.stack);
      }
    }
  }
}

// XenovaEmbeddings.ts
// Wrapper LangChain Embeddings utilisant @xenova/transformers (100% local, offline)
// Le modèle (~80MB) est téléchargé une seule fois dans ~/.cache/huggingface/hub/

import { Embeddings } from '@langchain/core/embeddings';
import { XENOVA_MODEL } from './config.js';
import { logger } from './logger.js';

// Pipeline chargé en lazy — une seule instance par process
let pipelineInstance: any = null;

async function getPipeline(modelName: string): Promise<any> {
  if (pipelineInstance === null) {
    logger.info(`Loading Xenova model: ${modelName} (first call triggers download if not cached)`);
    // Import dynamique requis : @xenova/transformers est ESM-only
    const { pipeline, env } = await import('@xenova/transformers');
    // Désactiver le progress bar (inutile en MCP stdio)
    env.useBrowserCache = false;
    pipelineInstance = await pipeline('feature-extraction', modelName);
    logger.info('Xenova model loaded successfully');
  }
  return pipelineInstance;
}

export class XenovaEmbeddings extends Embeddings {
  private modelName: string;

  constructor(modelName?: string) {
    super({});
    this.modelName = modelName ?? XENOVA_MODEL;
  }

  async embedDocuments(texts: string[]): Promise<number[][]> {
    const pipe = await getPipeline(this.modelName);
    const results: number[][] = [];
    for (const text of texts) {
      const output = await pipe(text, { pooling: 'mean', normalize: true });
      results.push(Array.from(output.data as Float32Array));
    }
    return results;
  }

  async embedQuery(text: string): Promise<number[]> {
    const pipe = await getPipeline(this.modelName);
    const output = await pipe(text, { pooling: 'mean', normalize: true });
    return Array.from(output.data as Float32Array);
  }
}

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import 'dotenv/config'; // Load .env
import { FETCH_NO_REDIRECT } from '../src/utils/fetch-safe.js';
import type { GitHubRepo } from './repo-scorer.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ROOT_DIR = path.resolve(__dirname, '..');
const KEYWORDS_FILE = path.join(ROOT_DIR, 'docs', 'optimization', 'keywords.json');
const DATA_DIR = path.join(ROOT_DIR, 'data');
const OUTPUT_FILE = path.join(DATA_DIR, 'raw_repos.json');

// Ensure data directory exists
if (!fs.existsSync(DATA_DIR)) {
  fs.mkdirSync(DATA_DIR, { recursive: true });
}

interface KeywordsConfig {
  categories: {
    [key: string]: {
      keywords: string[];
      tags: string[];
    };
  };
  criteria: {
    minStars: number;
    lastUpdatedDays: number;
    languages: string[];
  };
}

async function searchGitHub(query: string, token?: string) {
  const url = `https://api.github.com/search/repositories?q=${encodeURIComponent(query)}&sort=stars&order=desc&per_page=10`;
  const headers: Record<string, string> = {
    'Accept': 'application/vnd.github.v3+json',
    'User-Agent': 'OpenClaw-Optimization-Scanner',
  };
  
  if (token) {
    headers['Authorization'] = `token ${token}`;
  }

  try {
    const response = await fetch(url, { ...FETCH_NO_REDIRECT, headers });
    if (!response.ok) {
      console.error(`Failed to fetch for query "${query}": ${response.status} ${response.statusText}`);
      // Handle rate limits
      if (response.status === 403 || response.status === 429) {
          const resetTime = response.headers.get('x-ratelimit-reset');
          if (resetTime) {
            console.warn(`Rate limit exceeded. Reset at ${new Date(parseInt(resetTime) * 1000).toISOString()}`);
          }
      }
      return [];
    }
    const data = (await response.json()) as { items?: unknown[] };
    return (data.items ?? []) as Array<Record<string, unknown> & { id: number }>;
  } catch (error) {
    console.error(`Error searching for "${query}":`, error);
    return [];
  }
}

export interface ScannerOptions {
  limitPerCategory?: number;
  outputFile?: string;
  /** Delay between GitHub API calls (ms). Default 2000 for live runs; use 0 in tests. */
  rateLimitDelayMs?: number;
}

export async function scanGitHub(options: ScannerOptions = {}) {
  const rateLimitDelayMs =
    typeof options.rateLimitDelayMs === "number"
      ? options.rateLimitDelayMs
      : 2000;
  console.log('Starting GitHub Scanner...');
  
  // Read keywords
  if (!fs.existsSync(KEYWORDS_FILE)) {
    console.error(`Keywords file not found at ${KEYWORDS_FILE}`);
    throw new Error('Keywords file not found');
  }
  
  const config = JSON.parse(fs.readFileSync(KEYWORDS_FILE, 'utf-8')) as KeywordsConfig;
  const token = process.env.GITHUB_TOKEN;
  
  if (!token) {
    console.warn('No GITHUB_TOKEN found in environment variables. Rate limits will be strict.');
  }

  const allRepos: GitHubRepo[] = [];
  const seenIds = new Set<number>();
  const limit = options.limitPerCategory || 5;

  for (const [category, data] of Object.entries(config.categories)) {
    console.log(`Scanning category: ${category}`);
    
    // Construct queries
    // Strategy: Combine "keyword" + "language:typescript" (or other languages)
    // We limit queries to avoid hitting rate limits too fast
    const topKeywords = data.keywords.slice(0, limit); 
    
    for (const keyword of topKeywords) {
      const query = `${keyword} stars:>${config.criteria.minStars} pushed:>${getDateString(config.criteria.lastUpdatedDays)}`;
      console.log(`  Querying: ${query}`);
      
      const items = await searchGitHub(query, token);
      
      for (const item of items) {
        if (!seenIds.has(item.id)) {
          seenIds.add(item.id);
          allRepos.push({
            ...(item as unknown as GitHubRepo),
            _category: category,
            _scanned_at: new Date().toISOString(),
          });
        }
      }
      
      if (rateLimitDelayMs > 0) {
        await new Promise((resolve) => setTimeout(resolve, rateLimitDelayMs));
      }
    }
  }

  // Write results if outputFile is provided
  if (options.outputFile) {
    fs.writeFileSync(options.outputFile, JSON.stringify(allRepos, null, 2));
    console.log(`Saved ${allRepos.length} repositories to ${options.outputFile}`);
  }
  
  return allRepos;
}

// Only run main if executed directly
if (process.argv[1] === fileURLToPath(import.meta.url)) {
    scanGitHub({ outputFile: OUTPUT_FILE }).catch(console.error);
  }
  
  function getDateString(daysAgo: number): string {
    const date = new Date();
    date.setDate(date.getDate() - daysAgo);
    return date.toISOString().split('T')[0];
  }


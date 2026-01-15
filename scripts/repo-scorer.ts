import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ROOT_DIR = path.resolve(__dirname, '..');
const DATA_DIR = path.join(ROOT_DIR, 'data');
const INPUT_FILE = path.join(DATA_DIR, 'raw_repos.json');
const OUTPUT_FILE = path.join(DATA_DIR, 'scored_repos.json');

export interface GitHubRepo {
  id: number;
  name: string;
  full_name: string;
  html_url: string;
  description: string;
  stargazers_count: number;
  forks_count: number;
  open_issues_count: number;
  pushed_at: string;
  created_at: string;
  size: number;
  language: string;
  _category: string;
  _scanned_at: string;
}

export interface ScoredRepo extends GitHubRepo {
  _score: number;
  _score_details: {
    activity: number;
    community: number;
    growth: number;
  };
}

function calculateScore(repo: GitHubRepo): ScoredRepo {
  const now = new Date();
  const pushedDate = new Date(repo.pushed_at);
  const createdDate = new Date(repo.created_at);
  
  // 1. Recency Score (0-30 points)
  // Max score if updated within 7 days, 0 if > 90 days
  const daysSincePush = (now.getTime() - pushedDate.getTime()) / (1000 * 60 * 60 * 24);
  const recencyScore = Math.max(0, 30 - (daysSincePush / 3)); 

  // 2. Community Score (0-40 points)
  // Based on stars and forks
  // Assume 10k stars = max score
  const starScore = Math.min(20, (repo.stargazers_count / 10000) * 20);
  const forkScore = Math.min(20, (repo.forks_count / 2000) * 20);
  const communityScore = starScore + forkScore;

  // 3. Growth/Health (0-30 points)
  // Issues ratio (healthy if issues < 10% of stars? Maybe not accurate)
  // Just use size and longevity for now as a proxy for maturity
  const ageDays = (now.getTime() - createdDate.getTime()) / (1000 * 60 * 60 * 24);
  const maturityScore = Math.min(15, (ageDays / 365) * 5); // 5 points per year, max 15
  
  // Bonus for TypeScript
  const langBonus = (repo.language === 'TypeScript' || repo.language === 'JavaScript') ? 15 : 0;

  const totalScore = recencyScore + communityScore + maturityScore + langBonus;

  return {
    ...repo,
    _score: Math.round(totalScore * 100) / 100,
    _score_details: {
      activity: Math.round(recencyScore),
      community: Math.round(communityScore),
      growth: Math.round(maturityScore + langBonus)
    }
  };
}

export function scoreRepos(rawRepos: GitHubRepo[], outputFile?: string): ScoredRepo[] {
  console.log(`Processing ${rawRepos.length} repositories...`);

  const scoredRepos = rawRepos.map(calculateScore).sort((a, b) => b._score - a._score);

  if (outputFile) {
    fs.writeFileSync(outputFile, JSON.stringify(scoredRepos, null, 2));
    console.log(`Saved ${scoredRepos.length} scored repositories to ${outputFile}`);
  }
  
  return scoredRepos;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  console.log('Starting Repo Scorer...');
  
  if (!fs.existsSync(INPUT_FILE)) {
    console.error(`Input file not found at ${INPUT_FILE}`);
    process.exit(1);
  }

  const rawRepos = JSON.parse(fs.readFileSync(INPUT_FILE, 'utf-8')) as GitHubRepo[];
  const scored = scoreRepos(rawRepos, OUTPUT_FILE);

  // Print top 5
  console.log('\nTop 5 Repositories:');
  scored.slice(0, 5).forEach((repo, i) => {
    console.log(`${i + 1}. [${repo._score}] ${repo.full_name} (${repo._category})`);
  });
}

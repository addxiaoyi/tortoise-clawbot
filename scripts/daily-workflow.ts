import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { scanGitHub } from './github-scanner.js';
import { scoreRepos } from './repo-scorer.js';
import { ImprovementsManager, ImprovementCategory } from './improvements-manager.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ROOT_DIR = path.resolve(__dirname, '..');
const DATA_DIR = path.join(ROOT_DIR, 'data');
const DAILY_CANDIDATES_FILE = path.join(DATA_DIR, 'daily_candidates.json');

async function main() {
  console.log('=== Starting Daily Optimization Workflow ===');
  
  // 1. Scan GitHub
  console.log('\n[Step 1] Scanning GitHub...');
  // Use a slightly higher limit for daily scan to get fresh candidates
  const rawRepos = await scanGitHub({ limitPerCategory: 5 }); 
  
  // 2. Score Repos
  console.log('\n[Step 2] Scoring Repositories...');
  const scoredRepos = scoreRepos(rawRepos);
  
  // 3. Deduplicate and Select Candidates
  console.log('\n[Step 3] Selecting Candidates...');
  const manager = new ImprovementsManager();
  const existingRepos = new Set(manager.getAll().map(i => i.source_repo));
  
  const newCandidates = scoredRepos.filter(repo => !existingRepos.has(repo.full_name));
  
  console.log(`Found ${scoredRepos.length} scored repos, ${newCandidates.length} are new.`);
  
  // Take top 3 new candidates
  const topCandidates = newCandidates.slice(0, 3);
  
  // 4. Add to DB
  console.log('\n[Step 4] Updating Improvements Database...');
  const addedImprovements = [];
  
  for (const candidate of topCandidates) {
    const improvement = manager.add(
      candidate.full_name,
      candidate._category as ImprovementCategory,
      candidate._score,
      `Auto-detected from daily scan. Score: ${candidate._score}. Description: ${candidate.description}`
    );
    addedImprovements.push(improvement);
  }
  
  // 5. Generate Daily Report
  console.log('\n[Step 5] Generating Daily Report...');
  fs.writeFileSync(DAILY_CANDIDATES_FILE, JSON.stringify(addedImprovements, null, 2));
  
  console.log(`\n=== Workflow Complete ===`);
  console.log(`Added ${addedImprovements.length} new improvements to the pipeline.`);
  console.log(`Report saved to ${DAILY_CANDIDATES_FILE}`);
}

main().catch(console.error);

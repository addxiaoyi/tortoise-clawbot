import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ROOT_DIR = path.resolve(__dirname, '..');
const DB_FILE = path.join(ROOT_DIR, 'data', 'improvements.db.json');

export type ImprovementCategory = 'skill' | 'mcp' | 'prompt';
export type ImprovementStatus = 'pending' | 'in_progress' | 'completed' | 'rejected';

export interface Improvement {
  id: string;
  source_repo: string;
  category: ImprovementCategory;
  status: ImprovementStatus;
  impact_score: number;
  description?: string;
  created_at: string;
  updated_at: string;
}

export class ImprovementsManager {
  private dbPath: string;
  private improvements: Improvement[];

  constructor(dbPath: string = DB_FILE) {
    this.dbPath = dbPath;
    this.improvements = [];
    this.load();
  }

  private load() {
    if (fs.existsSync(this.dbPath)) {
      try {
        const data = fs.readFileSync(this.dbPath, 'utf-8');
        this.improvements = JSON.parse(data);
      } catch (error) {
        console.error('Error loading improvements DB:', error);
        this.improvements = [];
      }
    }
  }

  public save() {
    fs.writeFileSync(this.dbPath, JSON.stringify(this.improvements, null, 2));
  }

  public getAll(): Improvement[] {
    return this.improvements;
  }

  public findByRepo(repoName: string): Improvement | undefined {
    return this.improvements.find(i => i.source_repo === repoName);
  }

  public add(repoName: string, category: ImprovementCategory, score: number, description?: string): Improvement {
    const existing = this.findByRepo(repoName);
    if (existing) {
      console.log(`Repo ${repoName} already exists in improvements DB.`);
      return existing;
    }

    const newImprovement: Improvement = {
      id: `imp_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
      source_repo: repoName,
      category,
      status: 'pending',
      impact_score: score,
      description: description || `Potential optimization from ${repoName}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    this.improvements.push(newImprovement);
    this.save();
    console.log(`Added new improvement: ${repoName} (${category})`);
    return newImprovement;
  }

  public updateStatus(id: string, status: ImprovementStatus) {
    const imp = this.improvements.find(i => i.id === id);
    if (imp) {
      imp.status = status;
      imp.updated_at = new Date().toISOString();
      this.save();
    }
  }
}

// CLI usage
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const manager = new ImprovementsManager();
  const args = process.argv.slice(2);
  const command = args[0];

  if (command === 'list') {
    console.table(manager.getAll().map(i => ({
      id: i.id,
      repo: i.source_repo,
      cat: i.category,
      status: i.status,
      score: i.impact_score
    })));
  } else if (command === 'add' && args.length >= 4) {
    // node scripts/improvements-manager.ts add user/repo skill 95 "desc"
    manager.add(args[1], args[2] as ImprovementCategory, parseFloat(args[3]), args[4]);
  } else {
    console.log('Usage: node scripts/improvements-manager.ts [list|add <repo> <cat> <score> [desc]]');
  }
}

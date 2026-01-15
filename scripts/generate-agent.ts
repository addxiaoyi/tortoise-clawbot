import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import readline from 'node:readline';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT_DIR = path.resolve(__dirname, '..');
const TEMPLATE_PATH = path.join(ROOT_DIR, 'templates', 'agent-config.json');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

function ask(question: string): Promise<string> {
  return new Promise(resolve => rl.question(question, resolve));
}

async function main() {
  console.log('=== OpenClaw Agent Config Generator ===\n');

  const agentName = await ask('Agent Name (kebab-case): ') || 'my-agent';
  const agentDescription = await ask('Description: ') || 'A new OpenClaw agent';
  
  const enableMemory = (await ask('Enable Memory? (y/N): ')).toLowerCase() === 'y';
  const enableWebSearch = (await ask('Enable Web Search? (y/N): ')).toLowerCase() === 'y';
  
  console.log('\nSelect permissions (comma separated):');
  console.log('1. file:read');
  console.log('2. file:write');
  console.log('3. shell:exec');
  console.log('4. net:fetch');
  
  const permsInput = await ask('Permissions: ');
  const permsMap: Record<string, string> = {
    '1': 'file:read',
    '2': 'file:write',
    '3': 'shell:exec',
    '4': 'net:fetch'
  };
  
  const permissions = permsInput.split(',')
    .map(p => p.trim())
    .filter(p => permsMap[p] || p)
    .map(p => permsMap[p] || p);

  // Read template
  let template = fs.readFileSync(TEMPLATE_PATH, 'utf-8');
  
  // Replace variables
  template = template.replace('{{agentName}}', agentName);
  template = template.replace('{{agentDescription}}', agentDescription);
  template = template.replace('{{enableMemory}}', String(enableMemory));
  template = template.replace('{{enableWebSearch}}', String(enableWebSearch));
  
  // Handle permissions array (hacky string replacement for JSON array)
  const permsJson = JSON.stringify(permissions).slice(1, -1); // remove [ ]
  template = template.replace('"{{permissions}}"', permsJson || '');

  // Ensure output dir
  const outputDir = path.join(ROOT_DIR, 'generated-agents', agentName);
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }
  
  const outputPath = path.join(outputDir, 'package.json');
  fs.writeFileSync(outputPath, template);
  
  console.log(`\nGenerated agent config at: ${outputPath}`);
  rl.close();
}

main().catch(console.error);

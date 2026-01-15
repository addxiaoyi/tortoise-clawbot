import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest';
import { ConfigWatcher } from './config-watcher.js';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

describe('ConfigWatcher', () => {
  let tmpDir: string;
  let configFile: string;
  let watcher: ConfigWatcher;

  beforeEach(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'config-watcher-test-'));
    configFile = path.join(tmpDir, 'config.json');
    fs.writeFileSync(configFile, JSON.stringify({ version: 1 }));
  });

  afterEach(() => {
    if (watcher) {
      watcher.stop();
    }
    fs.rmSync(tmpDir, { recursive: true, force: true });
  });

  it('should load initial config', async () => {
    watcher = new ConfigWatcher({ configPath: configFile });
    await watcher.start();
    
    // Give it a moment to load
    await new Promise(resolve => setTimeout(resolve, 50));
    
    expect(watcher.getConfig()).toEqual({ version: 1 });
  });

  it('should detect changes', async () => {
    watcher = new ConfigWatcher({ configPath: configFile, debounceInterval: 10 });
    await watcher.start();
    
    const changePromise = new Promise(resolve => {
      watcher.on('change', (newConfig) => {
        resolve(newConfig);
      });
    });

    // Write new config
    fs.writeFileSync(configFile, JSON.stringify({ version: 2 }));

    const newConfig = await changePromise;
    expect(newConfig).toEqual({ version: 2 });
  });

  it('should handle invalid json', async () => {
    watcher = new ConfigWatcher({ configPath: configFile, debounceInterval: 10 });
    await watcher.start();

    const errorPromise = new Promise(resolve => {
      watcher.on('error', (err) => resolve(err));
    });

    fs.writeFileSync(configFile, '{ invalid json');

    const error = await errorPromise;
    expect(error).toBeDefined();
  });
});

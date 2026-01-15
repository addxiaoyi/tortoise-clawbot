/**
 * Memory Module
 * 
 * Hierarchical memory management
 */

import { AxiosInstance } from 'axios';
import { MemoryType } from './types';

export interface MemoryEntry {
  key: string;
  value: unknown;
  type?: MemoryType;
}

export interface MemoryStats {
  episodic_count: number;
  semantic_count: number;
  procedural_count: number;
}

export interface SearchResult {
  results: MemoryEntry[];
}

export class Memory {
  constructor(private http: AxiosInstance) {}
  
  /**
   * Store a memory entry
   */
  async store(key: string, value: unknown, type: MemoryType = 'episodic'): Promise<void> {
    await this.http.post('/api/v1/memory', {
      key,
      value,
      memory_type: type,
    });
  }
  
  /**
   * Get a memory entry
   */
  async get(key: string): Promise<MemoryEntry | null> {
    try {
      const response = await this.http.get<MemoryEntry>(`/api/v1/memory/${encodeURIComponent(key)}`);
      return response.data;
    } catch (error: unknown) {
      if ((error as { response?: { status?: number } })?.response?.status === 404) {
        return null;
      }
      throw error;
    }
  }
  
  /**
   * Delete a memory entry
   */
  async delete(key: string): Promise<void> {
    await this.http.delete(`/api/v1/memory/${encodeURIComponent(key)}`);
  }
  
  /**
   * Search memories
   */
  async search(query: string, limit = 10): Promise<MemoryEntry[]> {
    const response = await this.http.post<SearchResult>('/api/v1/memory/search', {
      query,
      limit,
    });
    return response.data.results;
  }
  
  /**
   * Get memory statistics
   */
  async stats(): Promise<MemoryStats> {
    const response = await this.http.get<MemoryStats>('/api/v1/memory');
    return response.data;
  }
  
  /**
   * Compact memory (cleanup expired entries)
   */
  async compact(): Promise<void> {
    await this.http.post('/api/v1/memory/compact');
  }
}

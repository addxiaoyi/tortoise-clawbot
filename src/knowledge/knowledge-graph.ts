/**
 * Knowledge Graph - 知识图谱
 * 用于实体关系提取和推理
 */

export interface Entity {
  id: string;
  type: EntityType;
  name: string;
  properties: Record<string, unknown>;
  embeddings?: number[];
}

export type EntityType = 
  | 'person'
  | 'organization'
  | 'location'
  | 'concept'
  | 'event'
  | 'document'
  | 'custom';

export interface Relation {
  id: string;
  source: string;    // 源实体ID
  target: string;    // 目标实体ID
  type: RelationType;
  weight: number;    // 关系权重
  properties: Record<string, unknown>;
}

export type RelationType =
  | 'knows'
  | 'works_at'
  | 'located_in'
  | 'part_of'
  | 'related_to'
  | 'created'
  | 'mentioned'
  | 'custom';

export interface GraphQuery {
  entityId?: string;
  entityType?: EntityType;
  relationType?: RelationType;
  depth?: number;
  limit?: number;
}

export interface SearchResult {
  entities: Entity[];
  relations: Relation[];
  score: number;
}

/**
 * KnowledgeGraph - 知识图谱
 */
export class KnowledgeGraph {
  private entities: Map<string, Entity> = new Map();
  private relations: Map<string, Relation> = new Map();
  private entityIndex: Map<EntityType, Set<string>> = new Map();
  private relationIndex: Map<string, Set<string>> = new Map(); // entityId -> relationIds

  constructor() {
    // 初始化索引
    Object.values(EntityType).forEach(type => {
      this.entityIndex.set(type, new Set());
    });
  }

  /**
   * 添加实体
   */
  addEntity(entity: Omit<Entity, 'id'>): Entity {
    const id = crypto.randomUUID();
    const newEntity: Entity = { ...entity, id };
    
    this.entities.set(id, newEntity);
    this.entityIndex.get(entity.type)?.add(id);
    
    return newEntity;
  }

  /**
   * 添加关系
   */
  addRelation(relation: Omit<Relation, 'id'>): Relation {
    const id = crypto.randomUUID();
    const newRelation: Relation = { ...relation, id };
    
    this.relations.set(id, newRelation);
    
    // 更新索引
    if (!this.relationIndex.has(relation.source)) {
      this.relationIndex.set(relation.source, new Set());
    }
    this.relationIndex.get(relation.source)!.add(id);
    
    if (!this.relationIndex.has(relation.target)) {
      this.relationIndex.set(relation.target, new Set());
    }
    this.relationIndex.get(relation.target)!.add(id);
    
    return newRelation;
  }

  /**
   * 获取实体
   */
  getEntity(id: string): Entity | null {
    return this.entities.get(id) || null;
  }

  /**
   * 查询实体
   */
  queryEntities(query: GraphQuery): Entity[] {
    let results = Array.from(this.entities.values());

    if (query.entityId) {
      const entity = this.entities.get(query.entityId);
      return entity ? [entity] : [];
    }

    if (query.entityType) {
      const typeIds = this.entityIndex.get(query.entityType);
      if (typeIds) {
        results = results.filter(e => typeIds.has(e.id));
      }
    }

    const limit = query.limit || 20;
    return results.slice(0, limit);
  }

  /**
   * 获取实体的关系
   */
  getRelations(entityId: string, type?: RelationType): Relation[] {
    const relationIds = this.relationIndex.get(entityId);
    if (!relationIds) return [];

    let results = Array.from(relationIds)
      .map(id => this.relations.get(id))
      .filter((r): r is Relation => r !== undefined);

    if (type) {
      results = results.filter(r => r.type === type);
    }

    return results;
  }

  /**
   * 遍历图（广度优先）
   */
  traverse(startId: string, depth: number = 2): {
    entities: Entity[];
    relations: Relation[];
  } {
    const visitedEntities = new Set<string>();
    const visitedRelations = new Set<string>();
    const queue: Array<{ id: string; currentDepth: number }> = [{ id: startId, currentDepth: 0 }];

    while (queue.length > 0) {
      const { id, currentDepth } = queue.shift()!;

      if (visitedEntities.has(id) || currentDepth > depth) continue;
      visitedEntities.add(id);

      // 获取相邻实体
      const relations = this.getRelations(id);
      for (const relation of relations) {
        visitedRelations.add(relation.id);

        if (!visitedEntities.has(relation.target)) {
          queue.push({ id: relation.target, currentDepth: currentDepth + 1 });
        }
        if (!visitedEntities.has(relation.source)) {
          queue.push({ id: relation.source, currentDepth: currentDepth + 1 });
        }
      }
    }

    return {
      entities: Array.from(visitedEntities).map(id => this.entities.get(id)!),
      relations: Array.from(visitedRelations).map(id => this.relations.get(id)!),
    };
  }

  /**
   * 查找路径
   */
  findPath(startId: string, endId: string): Relation[] | null {
    const visited = new Set<string>();
    const queue: Array<{ id: string; path: Relation[] }> = [
      { id: startId, path: [] }
    ];

    while (queue.length > 0) {
      const { id, path } = queue.shift()!;

      if (id === endId) return path;

      if (visited.has(id)) continue;
      visited.add(id);

      const relations = this.getRelations(id);
      for (const relation of relations) {
        if (!visited.has(relation.target)) {
          queue.push({
            id: relation.target,
            path: [...path, relation]
          });
        }
      }
    }

    return null;
  }

  /**
   * 实体搜索
   */
  searchEntities(query: string): SearchResult {
    const queryLower = query.toLowerCase();
    const matchedEntities: Array<{ entity: Entity; score: number }> = [];

    for (const entity of this.entities.values()) {
      let score = 0;

      // 名称匹配
      if (entity.name.toLowerCase().includes(queryLower)) {
        score += 0.8;
      }

      // 属性匹配
      for (const value of Object.values(entity.properties)) {
        if (String(value).toLowerCase().includes(queryLower)) {
          score += 0.4;
        }
      }

      if (score > 0) {
        matchedEntities.push({ entity, score });
      }
    }

    // 按分数排序
    matchedEntities.sort((a, b) => b.score - a.score);

    // 收集相关关系
    const entityIds = new Set(matchedEntities.map(m => m.entity.id));
    const relations = Array.from(this.relations.values())
      .filter(r => entityIds.has(r.source) || entityIds.has(r.target));

    return {
      entities: matchedEntities.map(m => m.entity),
      relations,
      score: matchedEntities[0]?.score || 0,
    };
  }

  /**
   * 从文本提取实体和关系
   */
  async extractFromText(text: string): Promise<{
    entities: Entity[];
    relations: Relation[];
  }> {
    const entities: Entity[] = [];
    const relations: Relation[] = [];

    // 简单命名实体识别（实际应用中应使用NLP模型）
    const personMatches = text.match(/[A-Z][a-z]+ [A-Z][a-z]+/g);
    const orgMatches = text.match(/[A-Z][a-z]+ (Inc|LLC|Corp|Company)/g);

    // 创建人员实体
    if (personMatches) {
      for (const name of new Set(personMatches)) {
        entities.push(this.addEntity({
          type: 'person',
          name,
          properties: { source: 'extracted' },
        }));
      }
    }

    // 创建组织实体
    if (orgMatches) {
      for (const name of new Set(orgMatches)) {
        entities.push(this.addEntity({
          type: 'organization',
          name,
          properties: { source: 'extracted' },
        }));
      }
    }

    return { entities, relations };
  }

  /**
   * 获取统计
   */
  getStats(): {
    totalEntities: number;
    totalRelations: number;
    entitiesByType: Record<EntityType, number>;
    relationsByType: Record<RelationType, number>;
  } {
    const entitiesByType: Record<EntityType, number> = {} as any;
    const relationsByType: Record<RelationType, number> = {} as any;

    for (const entity of this.entities.values()) {
      entitiesByType[entity.type] = (entitiesByType[entity.type] || 0) + 1;
    }

    for (const relation of this.relations.values()) {
      relationsByType[relation.type] = (relationsByType[relation.type] || 0) + 1;
    }

    return {
      totalEntities: this.entities.size,
      totalRelations: this.relations.size,
      entitiesByType,
      relationsByType,
    };
  }

  /**
   * 导出图数据
   */
  export(): { entities: Entity[]; relations: Relation[] } {
    return {
      entities: Array.from(this.entities.values()),
      relations: Array.from(this.relations.values()),
    };
  }

  /**
   * 导入图数据
   */
  async import(data: { entities: Entity[]; relations: Relation[] }): Promise<void> {
    for (const entity of data.entities) {
      this.entities.set(entity.id, entity);
      this.entityIndex.get(entity.type)?.add(entity.id);
    }

    for (const relation of data.relations) {
      this.relations.set(relation.id, relation);
      
      if (!this.relationIndex.has(relation.source)) {
        this.relationIndex.set(relation.source, new Set());
      }
      this.relationIndex.get(relation.source)!.add(relation.id);
      
      if (!this.relationIndex.has(relation.target)) {
        this.relationIndex.set(relation.target, new Set());
      }
      this.relationIndex.get(relation.target)!.add(relation.id);
    }
  }
}

// 导出单例
export const knowledgeGraph = new KnowledgeGraph();

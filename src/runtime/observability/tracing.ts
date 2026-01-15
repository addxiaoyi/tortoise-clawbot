/**
 * Observability System - 监控 + 追踪 + 指标
 */

import crypto from 'node:crypto';
import { EventEmitter } from 'node:events';

// ==================== 类型定义 ====================

export interface Trace {
  id: string;
  name: string;
  startTime: number;
  endTime?: number;
  duration?: number;
  status: 'running' | 'completed' | 'error';
  error?: string;
  metadata?: Record<string, unknown>;
  spans: Span[];
  events: TraceEvent[];
}

export interface Span {
  id: string;
  traceId: string;
  parentId?: string;
  name: string;
  startTime: number;
  endTime?: number;
  duration?: number;
  tags: Record<string, string | number | boolean>;
  logs: SpanLog[];
}

export interface SpanLog {
  timestamp: number;
  fields: Record<string, unknown>;
}

export interface TraceEvent {
  timestamp: number;
  name: string;
  fields: Record<string, unknown>;
}

export interface Metric {
  name: string;
  type: 'counter' | 'gauge' | 'histogram' | 'summary';
  value: number;
  timestamp: number;
  tags: Record<string, string>;
}

export interface MetricPoint {
  timestamp: number;
  value: number;
}

export interface HistogramBuckets {
  [key: number]: number;
}

export interface LogEntry {
  id: string;
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  message: string;
  timestamp: number;
  traceId?: string;
  spanId?: string;
  service: string;
  fields: Record<string, unknown>;
}

export interface Dashboard {
  id: string;
  name: string;
  widgets: DashboardWidget[];
  refreshIntervalMs: number;
  createdAt: number;
  updatedAt: number;
}

export interface DashboardWidget {
  type: 'chart' | 'gauge' | 'stat' | 'table' | 'log';
  title: string;
  metrics: string[];
  config: Record<string, unknown>;
}

// ==================== 追踪系统 ====================

export class Tracer extends EventEmitter {
  private traces = new Map<string, Trace>();
  private activeSpans = new Map<string, Span>();
  private maxTraces = 1000;
  private maxSpansPerTrace = 10000;

  constructor(private serviceName: string = 'agent-runtime') {
    super();
  }

  startTrace(name: string, metadata?: Record<string, unknown>): Trace {
    const id = crypto.randomUUID();
    const trace: Trace = {
      id,
      name,
      startTime: Date.now(),
      status: 'running',
      metadata,
      spans: [],
      events: [],
    };

    this.traces.set(id, trace);
    this.emit('trace:start', trace);

    if (this.traces.size > this.maxTraces) {
      const oldest = this.findOldestTrace();
      if (oldest) this.traces.delete(oldest);
    }

    return trace;
  }

  endTrace(traceId: string, error?: Error): Trace | undefined {
    const trace = this.traces.get(traceId);
    if (!trace) return undefined;

    trace.endTime = Date.now();
    trace.duration = trace.endTime - trace.startTime;
    trace.status = error ? 'error' : 'completed';
    if (error) trace.error = error.message;

    for (const span of this.activeSpans.values()) {
      if (span.traceId === traceId) this.endSpan(span.id);
    }

    this.emit('trace:end', trace);
    return trace;
  }

  startSpan(name: string, options?: { traceId?: string; parentId?: string; tags?: Record<string, string | number | boolean> }): Span {
    const traceId = options?.traceId || this.findActiveTraceId();
    const id = crypto.randomUUID();

    const span: Span = {
      id,
      traceId,
      parentId: options?.parentId,
      name,
      startTime: Date.now(),
      tags: options?.tags || {},
      logs: [],
    };

    this.activeSpans.set(id, span);
    this.emit('span:start', span);
    return span;
  }

  endSpan(spanId: string): Span | undefined {
    const span = this.activeSpans.get(spanId);
    if (!span) return undefined;

    span.endTime = Date.now();
    span.duration = span.endTime - span.startTime;

    const trace = this.traces.get(span.traceId);
    if (trace) {
      trace.spans.push(span);
      if (trace.spans.length > this.maxSpansPerTrace) {
        trace.spans = trace.spans.slice(-this.maxSpansPerTrace);
      }
    }

    this.activeSpans.delete(spanId);
    this.emit('span:end', span);
    return span;
  }

  setSpanTag(spanId: string, key: string, value: string | number | boolean): void {
    const span = this.activeSpans.get(spanId) || this.findSpanById(spanId);
    if (span) span.tags[key] = value;
  }

  logSpan(spanId: string, fields: Record<string, unknown>): void {
    const span = this.activeSpans.get(spanId) || this.findSpanById(spanId);
    if (span) span.logs.push({ timestamp: Date.now(), fields });
  }

  addTraceEvent(traceId: string, name: string, fields: Record<string, unknown> = {}): void {
    const trace = this.traces.get(traceId);
    if (trace) trace.events.push({ timestamp: Date.now(), name, fields });
  }

  getTrace(traceId: string): Trace | undefined {
    return this.traces.get(traceId);
  }

  listTraces(options?: { since?: number; status?: 'running' | 'completed' | 'error'; limit?: number }): Trace[] {
    let traces = Array.from(this.traces.values());

    if (options?.since) traces = traces.filter(t => t.startTime >= options.since!);
    if (options?.status) traces = traces.filter(t => t.status === options.status);
    traces.sort((a, b) => b.startTime - a.startTime);
    if (options?.limit) traces = traces.slice(0, options.limit);

    return traces;
  }

  private findActiveTraceId(): string | undefined {
    for (const span of this.activeSpans.values()) return span.traceId;
    return undefined;
  }

  private findOldestTrace(): string | undefined {
    let oldest: string | undefined;
    let oldestTime = Infinity;
    for (const [id, trace] of this.traces) {
      if (trace.startTime < oldestTime) { oldestTime = trace.startTime; oldest = id; }
    }
    return oldest;
  }

  private findSpanById(spanId: string): Span | undefined {
    for (const trace of this.traces.values()) {
      const span = trace.spans.find(s => s.id === spanId);
      if (span) return span;
    }
    return undefined;
  }
}

// ==================== 指标系统 ====================

export class Metrics extends EventEmitter {
  private counters = new Map<string, { value: number; tags: Record<string, string> }>();
  private gauges = new Map<string, { value: number; tags: Record<string, string> }>();
  private histograms = new Map<string, { values: number[]; buckets: HistogramBuckets; tags: Record<string, string> }>();
  private series = new Map<string, MetricPoint[]>();
  private maxPointsPerSeries = 1000;

  incrementCounter(name: string, value: number = 1, tags: Record<string, string> = {}): void {
    const key = this.makeKey(name, tags);
    const existing = this.counters.get(key);
    if (existing) existing.value += value;
    else this.counters.set(key, { value, tags });
    this.emit('metric', { name, type: 'counter', value: this.counters.get(key)!.value, tags, timestamp: Date.now() });
  }

  setGauge(name: string, value: number, tags: Record<string, string> = {}): void {
    this.gauges.set(this.makeKey(name, tags), { value, tags });
    this.emit('metric', { name, type: 'gauge', value, tags, timestamp: Date.now() });
  }

  recordHistogram(name: string, value: number, tags: Record<string, string> = {}): void {
    const key = this.makeKey(name, tags);
    const existing = this.histograms.get(key);
    if (existing) {
      existing.values.push(value);
      this.updateBuckets(existing.buckets, value);
    } else {
      const buckets: HistogramBuckets = {};
      this.updateBuckets(buckets, value);
      this.histograms.set(key, { values: [value], buckets, tags });
    }
    this.emit('metric', { name, type: 'histogram', value, tags, timestamp: Date.now() });
  }

  getHistogramStats(name: string, tags: Record<string, string> = {}): { count: number; sum: number; min: number; max: number; avg: number; p50: number; p90: number; p99: number } {
    const key = this.makeKey(name, tags);
    const hist = this.histograms.get(key);
    if (!hist || hist.values.length === 0) return { count: 0, sum: 0, min: 0, max: 0, avg: 0, p50: 0, p90: 0, p99: 0 };
    const sorted = [...hist.values].sort((a, b) => a - b);
    const sum = sorted.reduce((a, b) => a + b, 0);
    return { count: sorted.length, sum, min: sorted[0], max: sorted[sorted.length - 1], avg: sum / sorted.length, p50: this.percentile(sorted, 0.5), p90: this.percentile(sorted, 0.9), p99: this.percentile(sorted, 0.99) };
  }

  recordSeries(name: string, value: number, tags: Record<string, string> = {}): void {
    const key = this.makeKey(name, tags);
    const point: MetricPoint = { timestamp: Date.now(), value };
    const existing = this.series.get(key);
    if (existing) { existing.push(point); if (existing.length > this.maxPointsPerSeries) existing.shift(); }
    else this.series.set(key, [point]);
  }

  collect(): Metric[] {
    const metrics: Metric[] = [];
    const now = Date.now();
    for (const [, { value, tags }] of this.counters) metrics.push({ name: 'counter', type: 'counter', value, timestamp: now, tags });
    for (const [, { value, tags }] of this.gauges) metrics.push({ name: 'gauge', type: 'gauge', value, timestamp: now, tags });
    for (const [key, hist] of this.histograms) metrics.push({ name: 'histogram', type: 'histogram', value: hist.values.length, timestamp: now, tags: hist.tags });
    return metrics;
  }

  private makeKey(name: string, tags: Record<string, string>): string {
    const tagStr = Object.entries(tags).sort(([a], [b]) => a.localeCompare(b)).map(([k, v]) => `${k}=${v}`).join(',');
    return tagStr ? `${name}{${tagStr}}` : name;
  }

  private updateBuckets(buckets: HistogramBuckets, value: number): void {
    for (const boundary of [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]) {
      if (value <= boundary) buckets[boundary] = (buckets[boundary] || 0) + 1;
    }
    buckets[Infinity] = (buckets[Infinity] || 0) + 1;
  }

  private percentile(sorted: number[], p: number): number {
    return sorted[Math.max(0, Math.ceil(sorted.length * p) - 1)];
  }
}

// ==================== 日志系统 ====================

export class Logger extends EventEmitter {
  private logs: LogEntry[] = [];
  private maxLogs = 10000;
  private serviceName: string;

  constructor(serviceName: string = 'agent-runtime') {
    super();
    this.serviceName = serviceName;
  }

  log(level: LogEntry['level'], message: string, fields: Record<string, unknown> = {}, traceId?: string, spanId?: string): LogEntry {
    const entry: LogEntry = { id: crypto.randomUUID(), level, message, timestamp: Date.now(), traceId, spanId, service: this.serviceName, fields };
    this.logs.push(entry);
    if (this.logs.length > this.maxLogs) this.logs = this.logs.slice(-this.maxLogs);
    this.emit('log', entry);
    if (level === 'error' || level === 'fatal') console.error(`[${level.toUpperCase()}] ${message}`, fields);
    else console.log(`[${level.toUpperCase()}] ${message}`, fields);
    return entry;
  }

  debug(message: string, fields?: Record<string, unknown>) { return this.log('debug', message, fields); }
  info(message: string, fields?: Record<string, unknown>) { return this.log('info', message, fields); }
  warn(message: string, fields?: Record<string, unknown>) { return this.log('warn', message, fields); }
  error(message: string, fields?: Record<string, unknown>) { return this.log('error', message, fields); }
  fatal(message: string, fields?: Record<string, unknown>) { return this.log('fatal', message, fields); }

  query(options?: { since?: number; until?: number; level?: LogEntry['level']; search?: string; traceId?: string; limit?: number }): LogEntry[] {
    let results = [...this.logs];
    if (options?.since) results = results.filter(l => l.timestamp >= options.since!);
    if (options?.until) results = results.filter(l => l.timestamp <= options.until!);
    if (options?.level) results = results.filter(l => l.level === options.level);
    if (options?.search) results = results.filter(l => l.message.toLowerCase().includes(options.search!.toLowerCase()));
    if (options?.traceId) results = results.filter(l => l.traceId === options.traceId);
    results.sort((a, b) => b.timestamp - a.timestamp);
    if (options?.limit) results = results.slice(0, options.limit);
    return results;
  }
}

// ==================== 观测性管理器 ====================

export class ObservabilityManager {
  tracer: Tracer;
  metrics: Metrics;
  logger: Logger;

  constructor(serviceName: string = 'agent-runtime') {
    this.tracer = new Tracer(serviceName);
    this.metrics = new Metrics();
    this.logger = new Logger(serviceName);
  }

  startOperation(name: string, metadata?: Record<string, unknown>) {
    const trace = this.tracer.startTrace(name, metadata);
    let currentSpan: Span | undefined;

    return {
      trace,
      end: (error?: Error) => {
        if (currentSpan) this.tracer.endSpan(currentSpan.id);
        this.tracer.endTrace(trace.id, error);
      },
      span: (name: string, tags?: Record<string, string | number | boolean>) => {
        if (currentSpan) this.tracer.endSpan(currentSpan.id);
        currentSpan = this.tracer.startSpan(name, { traceId: trace.id, tags });
        this.tracer.addTraceEvent(trace.id, `span:${name}:start`, { name });
        return {
          end: () => { this.tracer.endSpan(currentSpan!.id); currentSpan = undefined; },
          tag: (key: string, value: string | number | boolean) => this.tracer.setSpanTag(currentSpan!.id, key, value),
          log: (fields: Record<string, unknown>) => this.tracer.logSpan(currentSpan!.id, fields),
        };
      },
    };
  }

  async healthCheck() {
    const traces = this.tracer.listTraces({ status: 'running' });
    return {
      healthy: true,
      checks: {
        tracer: { status: 'ok', message: `${this.tracer.listTraces().length} traces, ${traces.length} active` },
        metrics: { status: 'ok', message: `${this.metrics.collect().length} metrics` },
        logging: { status: 'ok', message: `${this.logger.query().length} logs` },
      },
    };
  }

  toPrometheus(): string {
    return this.metrics.collect().map(m => {
      const tags = Object.entries(m.tags).map(([k, v]) => `${k}="${v}"`).join(',');
      const labels = tags ? `{${tags}}` : '';
      return `${m.name}${labels} ${m.value}`;
    }).join('\n');
  }
}

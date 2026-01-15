import { describe, it, expect } from 'vitest';
import { Container } from './container.js';

describe('Container', () => {
  it('should register and resolve instances', () => {
    const container = new Container();
    const service = { foo: 'bar' };
    
    container.register('my-service', service);
    expect(container.resolve('my-service')).toBe(service);
  });

  it('should register and resolve factories (singleton)', () => {
    const container = new Container();
    let count = 0;
    const factory = () => ({ id: ++count });

    container.registerFactory('factory-service', factory);

    const instance1 = container.resolve<any>('factory-service');
    const instance2 = container.resolve<any>('factory-service');

    expect(instance1.id).toBe(1);
    expect(instance2.id).toBe(1); // Should be same instance
    expect(instance1).toBe(instance2);
  });

  it('should throw for missing service', () => {
    const container = new Container();
    expect(() => container.resolve('missing')).toThrow('Service not registered: missing');
  });

  it('should check existence', () => {
    const container = new Container();
    container.register('exists', {});
    
    expect(container.has('exists')).toBe(true);
    expect(container.has('missing')).toBe(false);
  });
});

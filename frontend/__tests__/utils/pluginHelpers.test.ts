import { describe, it, expect } from 'vitest';
import { resolvePrerequisiteTask } from "@/utils/utils";

describe('pluginHelpers', () => {
  it('correctly formats task target and version for PHP', () => {
    const task = {
      name: 'PHP',
      version: '8.1',
      versionUrls: { '8.2': 'http://example.com/php82.zip' },
    };
    const resolved = resolvePrerequisiteTask(task, '8.2');

    expect(resolved.version).toBe('8.2');
    expect(resolved.target).toBe('php/php-8.2');
    expect(resolved.url).toBe('http://example.com/php82.zip');
  });

  it('correctly maps special category for Node.js and HeidiSQL', () => {
    const nodeTask = { name: 'Node.js', version: '18.0.0' };
    expect(resolvePrerequisiteTask(nodeTask).target).toBe('nodejs/node-v18.0.0');

    const heidiTask = { name: 'HeidiSQL', version: '12.0' };
    expect(resolvePrerequisiteTask(heidiTask).target).toBe('heidisql/12.0');
  });

  it('falls back to default task version if selectedVersion is omitted', () => {
    const apacheTask = { name: 'Apache', version: '2.4' };
    const resolved = resolvePrerequisiteTask(apacheTask);
    expect(resolved.version).toBe('2.4');
    expect(resolved.target).toBe('apache/httpd-2.4');
  });
});

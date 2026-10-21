export function resolvePrerequisiteTask(task: any, selectedVersion?: string) {
  const selectedVer = selectedVersion || task.version;
  const prefixMap: Record<string, string> = {
    'PHP': 'php-',
    'Apache': 'httpd-',
    'MySQL': 'mysql-',
    'Nginx': 'nginx-',
    'OpenSSL': 'openssl-',
    'Node.js': 'node-v',
    'Python': 'python-',
  };
  const categoryMap: Record<string, string> = {
    'Node.js': 'nodejs',
    'HeidiSQL': 'heidisql',
  };
  const category = categoryMap[task.name] || task.name.toLowerCase();
  const prefix = prefixMap[task.name] || '';

  const modifiedTask = {
    ...task,
    version: selectedVer,
    target: `${category}/${prefix}${selectedVer}`,
  };

  if (task.versionUrls?.[selectedVer]) {
    modifiedTask.url = task.versionUrls[selectedVer];
  }

  return modifiedTask;
}

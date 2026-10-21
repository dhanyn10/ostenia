export interface ServiceInfo {
  name: string;
  status: string;
  pid: number;
  port: number;
  ports: number[];
  activeVersion: string;
  remainingDays?: number;
}

export const INITIAL_SERVICES: ServiceInfo[] = [
  { name: 'Apache', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
  { name: 'Nginx', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
  { name: 'MySQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
  { name: 'PHP', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
  { name: 'Node.js', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
  { name: 'Python', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
  { name: 'HeidiSQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
  { name: 'OpenSSL', status: 'Stopped', pid: 0, port: 0, remainingDays: 0, ports: [], activeVersion: '' },
];

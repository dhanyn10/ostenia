import * as AppBackend from '../../wailsjs/go/backend/App';

export const AppService = {
  // App & Window
  quit: () => AppBackend.Quit(),
  openFileDialog: (options: any) => AppBackend.OpenFileDialog(options),
  openDirectoryDialog: (options: any) => AppBackend.OpenDirectoryDialog(options),
  saveFileDialog: (options: any) => AppBackend.SaveFileDialog(options),

  // Config
  getConfig: () => (AppBackend as any).GetConfig ? (AppBackend as any).GetConfig() : Promise.resolve({}),
  saveConfig: (cfg: any) => (AppBackend as any).SaveConfig ? (AppBackend as any).SaveConfig(cfg) : Promise.resolve(),
  getBaseDir: () => (AppBackend as any).GetBaseDir ? (AppBackend as any).GetBaseDir() : Promise.resolve(''),

  // Services
  getServices: () => (AppBackend as any).GetServices ? (AppBackend as any).GetServices() : Promise.resolve([]),
  startService: (name: string) => AppBackend.StartService(name),
  stopService: (name: string) => AppBackend.StopService(name),
  switchVersion: (serviceName: string, version: string) => (AppBackend as any).SwitchVersion ? (AppBackend as any).SwitchVersion(serviceName, version) : Promise.resolve(),
  setApacheHTTPS: (enable: boolean) => AppBackend.SetApacheHTTPS(enable),
  setNginxHTTPS: (enable: boolean) => AppBackend.SetNginxHTTPS(enable),

  // Plugins
  getPlugins: () => (AppBackend as any).GetPlugins ? (AppBackend as any).GetPlugins() : Promise.resolve([]),
  installPlugin: (name: string) => (AppBackend as any).InstallPlugin ? (AppBackend as any).InstallPlugin(name) : Promise.resolve(),
  installPrerequisite: (task: any) => AppBackend.InstallPrerequisite(task),
  deleteVersion: (category: string, version: string) => AppBackend.DeleteVersion(category, version),
  processCustomVersion: (serviceName: string, sourcePath: string) => (AppBackend as any).ProcessCustomVersion ? (AppBackend as any).ProcessCustomVersion(serviceName, sourcePath) : Promise.resolve(),
  installPluginModule: (parentName: string, moduleName: string) => AppBackend.InstallPluginModule(parentName, moduleName),
  uninstallPluginModule: (parentName: string, moduleName: string) => AppBackend.UninstallPluginModule(parentName, moduleName),

  // SSL
  generateRootCA: (destDir: string) => AppBackend.GenerateRootCA(destDir),
  getRemainingDays: (certPath: string) => AppBackend.GetRemainingDays(certPath),

  // SSH
  getSSHSessions: () => AppBackend.GetSSHSessions(),
  saveSSHProfile: (profile: any) => AppBackend.SaveSSHProfile(profile),
  deleteSSHProfile: (id: string) => AppBackend.DeleteSSHProfile(id),
  connectSSH: (profile: any) => AppBackend.ConnectSSH(profile),
  getRemoteCurrentPath: (profile: any) => AppBackend.GetRemoteCurrentPath(profile),
  executeSFTPAction: (profile: any, action: string, remotePath: string, extraArg: string) =>
    AppBackend.ExecuteSFTPAction(profile, action, remotePath, extraArg),
  getWSLDistros: () => AppBackend.GetWSLDistros(),
  getSSHResourceUsage: (profile: any) => AppBackend.GetSSHResourceUsage(profile),
};

export namespace backend {

	export class InstalledApp {
	    name: string;
	    path: string;

	    static createFrom(source: any = {}) {
	        return new InstalledApp(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class ProxyAppInfo {
	    name: string;
	    port: number;

	    static createFrom(source: any = {}) {
	        return new ProxyAppInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.port = source["port"];
	    }
	}
	export class ProxyStatusInfo {
	    name: string;
	    isUp: boolean;
	    port: number;

	    static createFrom(source: any = {}) {
	        return new ProxyStatusInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isUp = source["isUp"];
	        this.port = source["port"];
	    }
	}

}

export namespace config {

	export class Config {
	    baseDir: string;
	    wwwRoot: string;
	    phpVersion: string;
	    nodeVersion: string;
	    apacheHttps: boolean;
	    nginxHttps: boolean;
	    proxies: Record<string, number>;
	    defaultEditor: string;

	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseDir = source["baseDir"];
	        this.wwwRoot = source["wwwRoot"];
	        this.phpVersion = source["phpVersion"];
	        this.nodeVersion = source["nodeVersion"];
	        this.apacheHttps = source["apacheHttps"];
	        this.nginxHttps = source["nginxHttps"];
	        this.proxies = source["proxies"];
	        this.defaultEditor = source["defaultEditor"];
	    }
	}
	export class SSHSession {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    user: string;
	    authMethod: string;
	    password?: string;
	    keyPath?: string;
	    passphrase?: string;
	    lastPath?: string;
	    createdAt: number;

	    static createFrom(source: any = {}) {
	        return new SSHSession(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.authMethod = source["authMethod"];
	        this.password = source["password"];
	        this.keyPath = source["keyPath"];
	        this.passphrase = source["passphrase"];
	        this.lastPath = source["lastPath"];
	        this.createdAt = source["createdAt"];
	    }
	}

}

export namespace plugins {

	export class PluginModule {
	    name: string;
	    isInstalled: boolean;
	    status: string;
	    version: string;

	    static createFrom(source: any = {}) {
	        return new PluginModule(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isInstalled = source["isInstalled"];
	        this.status = source["status"];
	        this.version = source["version"];
	    }
	}
	export class DownloadTask {
	    name: string;
	    url: string;
	    version: string;
	    versions: string[];
	    versionUrls: Record<string, string>;
	    installedVers: string[];
	    target: string;
	    checkFile: string;
	    isInstalled: boolean;
	    iconSvg: string;
	    info: string;
	    modules: PluginModule[];

	    static createFrom(source: any = {}) {
	        return new DownloadTask(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.version = source["version"];
	        this.versions = source["versions"];
	        this.versionUrls = source["versionUrls"];
	        this.installedVers = source["installedVers"];
	        this.target = source["target"];
	        this.checkFile = source["checkFile"];
	        this.isInstalled = source["isInstalled"];
	        this.iconSvg = source["iconSvg"];
	        this.info = source["info"];
	        this.modules = this.convertValues(source["modules"], PluginModule);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace service {

	export class PHPExtensionInfo {
	    name: string;
	    enabled: boolean;

	    static createFrom(source: any = {}) {
	        return new PHPExtensionInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	    }
	}
	export class RemoteFile {
	    name: string;
	    size: number;
	    isDir: boolean;
	    modTime: number;
	    mode: string;

	    static createFrom(source: any = {}) {
	        return new RemoteFile(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.isDir = source["isDir"];
	        this.modTime = source["modTime"];
	        this.mode = source["mode"];
	    }
	}
	export class ServiceDetailedInfo {
	    name: string;
	    status: string;
	    pid: number;
	    port: number;
	    ports: number[];
	    remainingDays?: number;
	    activeVersion?: string;

	    static createFrom(source: any = {}) {
	        return new ServiceDetailedInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.pid = source["pid"];
	        this.port = source["port"];
	        this.ports = source["ports"];
	        this.remainingDays = source["remainingDays"];
	        this.activeVersion = source["activeVersion"];
	    }
	}

}

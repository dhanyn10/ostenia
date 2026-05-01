export namespace config {
	
	export class Config {
	    baseDir: string;
	    wwwRoot: string;
	    phpVersion: string;
	    nodeVersion: string;
	    apacheHttps: boolean;
	    nginxHttps: boolean;
	    proxies: Record<string, number>;
	
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
	    }
	}

}

export namespace main {
	
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

export namespace plugins {
	
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


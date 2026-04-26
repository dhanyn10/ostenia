export namespace config {
	
	export class Config {
	    wwwRoot: string;
	    phpVersion: string;
	    nodeVersion: string;
	    apacheHttps: boolean;
	    nginxHttps: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.wwwRoot = source["wwwRoot"];
	        this.phpVersion = source["phpVersion"];
	        this.nodeVersion = source["nodeVersion"];
	        this.apacheHttps = source["apacheHttps"];
	        this.nginxHttps = source["nginxHttps"];
	    }
	}

}

export namespace download {
	
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
	    }
	}

}

export namespace service {
	
	export class ServiceDetailedInfo {
	    name: string;
	    status: string;
	    pid: number;
	    port: number;
	    ports: number[];
	    remainingDays?: number;
	
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
	    }
	}

}


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


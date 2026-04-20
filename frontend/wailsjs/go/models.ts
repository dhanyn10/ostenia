export namespace download {
	
	export class DownloadTask {
	    name: string;
	    url: string;
	    version: string;
	    target: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.version = source["version"];
	        this.target = source["target"];
	    }
	}

}


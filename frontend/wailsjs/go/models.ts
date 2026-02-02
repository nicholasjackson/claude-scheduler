export namespace db {
	
	export class Job {
	    id: string;
	    name: string;
	    startDate: string;
	    intervalValue: number;
	    intervalUnit: string;
	    prompt: string;
	    active: boolean;
	    nextRun: string;
	    lastRun: string;
	    status: string;
	    output: string;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.startDate = source["startDate"];
	        this.intervalValue = source["intervalValue"];
	        this.intervalUnit = source["intervalUnit"];
	        this.prompt = source["prompt"];
	        this.active = source["active"];
	        this.nextRun = source["nextRun"];
	        this.lastRun = source["lastRun"];
	        this.status = source["status"];
	        this.output = source["output"];
	    }
	}

}


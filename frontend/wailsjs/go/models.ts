export namespace api {
	
	export class AddDownloadRequest {
	    url: string;
	    filename?: string;
	    file_path?: string;
	    queue_id?: number;
	    max_connections?: number;
	    speed_limit?: number;
	    headers?: Record<string, string>;
	    user_agent?: string;
	    referrer?: string;
	    priority?: number;
	
	    static createFrom(source: any = {}) {
	        return new AddDownloadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.filename = source["filename"];
	        this.file_path = source["file_path"];
	        this.queue_id = source["queue_id"];
	        this.max_connections = source["max_connections"];
	        this.speed_limit = source["speed_limit"];
	        this.headers = source["headers"];
	        this.user_agent = source["user_agent"];
	        this.referrer = source["referrer"];
	        this.priority = source["priority"];
	    }
	}
	export class AddDownloadResponse {
	    success: boolean;
	    error?: string;
	    download?: models.DownloadItem;
	    download_id?: number;
	
	    static createFrom(source: any = {}) {
	        return new AddDownloadResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.download = this.convertValues(source["download"], models.DownloadItem);
	        this.download_id = source["download_id"];
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
	export class CreateQueueRequest {
	    name: string;
	    max_concurrent?: number;
	    speed_limit?: number;
	    auto_start?: boolean;
	    schedule_enabled?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CreateQueueRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.max_concurrent = source["max_concurrent"];
	        this.speed_limit = source["speed_limit"];
	        this.auto_start = source["auto_start"];
	        this.schedule_enabled = source["schedule_enabled"];
	    }
	}
	export class CreateQueueResponse {
	    success: boolean;
	    error?: string;
	    queue?: models.QueueModel;
	
	    static createFrom(source: any = {}) {
	        return new CreateQueueResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.queue = this.convertValues(source["queue"], models.QueueModel);
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
	export class DownloadInfoResponse {
	    success: boolean;
	    error?: string;
	    url: string;
	    filename: string;
	    content_length: number;
	    supports_range: boolean;
	    content_type: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadInfoResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.url = source["url"];
	        this.filename = source["filename"];
	        this.content_length = source["content_length"];
	        this.supports_range = source["supports_range"];
	        this.content_type = source["content_type"];
	    }
	}
	export class DownloadStats {
	    total_downloads: number;
	    active_downloads: number;
	    completed_downloads: number;
	    failed_downloads: number;
	    paused_downloads: number;
	    total_bytes: number;
	    downloaded_bytes: number;
	    total_speed: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloadStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_downloads = source["total_downloads"];
	        this.active_downloads = source["active_downloads"];
	        this.completed_downloads = source["completed_downloads"];
	        this.failed_downloads = source["failed_downloads"];
	        this.paused_downloads = source["paused_downloads"];
	        this.total_bytes = source["total_bytes"];
	        this.downloaded_bytes = source["downloaded_bytes"];
	        this.total_speed = source["total_speed"];
	    }
	}
	export class QueueStats {
	    queue_id: number;
	    queue_name: string;
	    is_running: boolean;
	    total_count: number;
	    active_count: number;
	    pending_count: number;
	    downloading_count: number;
	    completed_count: number;
	    failed_count: number;
	    paused_count: number;
	    canceled_count: number;
	    total_bytes: number;
	    downloaded_bytes: number;
	    overall_progress: number;
	
	    static createFrom(source: any = {}) {
	        return new QueueStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.queue_id = source["queue_id"];
	        this.queue_name = source["queue_name"];
	        this.is_running = source["is_running"];
	        this.total_count = source["total_count"];
	        this.active_count = source["active_count"];
	        this.pending_count = source["pending_count"];
	        this.downloading_count = source["downloading_count"];
	        this.completed_count = source["completed_count"];
	        this.failed_count = source["failed_count"];
	        this.paused_count = source["paused_count"];
	        this.canceled_count = source["canceled_count"];
	        this.total_bytes = source["total_bytes"];
	        this.downloaded_bytes = source["downloaded_bytes"];
	        this.overall_progress = source["overall_progress"];
	    }
	}
	export class QueueStatsResponse {
	    success: boolean;
	    error?: string;
	    stats?: QueueStats;
	
	    static createFrom(source: any = {}) {
	        return new QueueStatsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.stats = this.convertValues(source["stats"], QueueStats);
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
	export class UpdateQueueRequest {
	    queue_id: number;
	    name?: string;
	    max_concurrent?: number;
	    speed_limit?: number;
	    auto_start?: boolean;
	    schedule_enabled?: boolean;
	    schedule_start?: string;
	    schedule_stop?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateQueueRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.queue_id = source["queue_id"];
	        this.name = source["name"];
	        this.max_concurrent = source["max_concurrent"];
	        this.speed_limit = source["speed_limit"];
	        this.auto_start = source["auto_start"];
	        this.schedule_enabled = source["schedule_enabled"];
	        this.schedule_start = source["schedule_start"];
	        this.schedule_stop = source["schedule_stop"];
	    }
	}

}

export namespace config {
	
	export class Settings {
	    default_download_path: string;
	    max_concurrent_downloads: number;
	    default_connections: number;
	    global_speed_limit: number;
	    default_speed_limit: number;
	    min_part_size: number;
	    max_retries: number;
	    retry_delay: number;
	    user_agent: string;
	    connection_timeout: number;
	    read_timeout: number;
	    theme: string;
	    language: string;
	    show_notifications: boolean;
	    show_tray_icon: boolean;
	    minimize_to_tray: boolean;
	    start_minimized: boolean;
	    data_directory: string;
	    auto_backup: boolean;
	    backup_interval: number;
	    max_backups: number;
	    enable_logging: boolean;
	    log_level: string;
	    log_directory: string;
	    check_for_updates: boolean;
	    send_usage_stats: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.default_download_path = source["default_download_path"];
	        this.max_concurrent_downloads = source["max_concurrent_downloads"];
	        this.default_connections = source["default_connections"];
	        this.global_speed_limit = source["global_speed_limit"];
	        this.default_speed_limit = source["default_speed_limit"];
	        this.min_part_size = source["min_part_size"];
	        this.max_retries = source["max_retries"];
	        this.retry_delay = source["retry_delay"];
	        this.user_agent = source["user_agent"];
	        this.connection_timeout = source["connection_timeout"];
	        this.read_timeout = source["read_timeout"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	        this.show_notifications = source["show_notifications"];
	        this.show_tray_icon = source["show_tray_icon"];
	        this.minimize_to_tray = source["minimize_to_tray"];
	        this.start_minimized = source["start_minimized"];
	        this.data_directory = source["data_directory"];
	        this.auto_backup = source["auto_backup"];
	        this.backup_interval = source["backup_interval"];
	        this.max_backups = source["max_backups"];
	        this.enable_logging = source["enable_logging"];
	        this.log_level = source["log_level"];
	        this.log_directory = source["log_directory"];
	        this.check_for_updates = source["check_for_updates"];
	        this.send_usage_stats = source["send_usage_stats"];
	    }
	}

}

export namespace models {
	
	export class Progress {
	    bytes_downloaded: number;
	    total_bytes: number;
	    speed: number;
	    eta: number;
	    percentage: number;
	    // Go type: time
	    last_update: any;
	
	    static createFrom(source: any = {}) {
	        return new Progress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bytes_downloaded = source["bytes_downloaded"];
	        this.total_bytes = source["total_bytes"];
	        this.speed = source["speed"];
	        this.eta = source["eta"];
	        this.percentage = source["percentage"];
	        this.last_update = this.convertValues(source["last_update"], null);
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
	export class DownloadItem {
	    id: number;
	    url: string;
	    filename: string;
	    file_path: string;
	    total_size: number;
	    status: number;
	    progress: Progress;
	    parts: number[];
	    max_connections: number;
	    speed_limit: number;
	    headers: Record<string, string>;
	    user_agent: string;
	    referrer: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    started_at?: any;
	    // Go type: time
	    completed_at?: any;
	    last_error?: string;
	    retry_count: number;
	    max_retries: number;
	    queue_id: number;
	    priority: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloadItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.filename = source["filename"];
	        this.file_path = source["file_path"];
	        this.total_size = source["total_size"];
	        this.status = source["status"];
	        this.progress = this.convertValues(source["progress"], Progress);
	        this.parts = source["parts"];
	        this.max_connections = source["max_connections"];
	        this.speed_limit = source["speed_limit"];
	        this.headers = source["headers"];
	        this.user_agent = source["user_agent"];
	        this.referrer = source["referrer"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
	        this.last_error = source["last_error"];
	        this.retry_count = source["retry_count"];
	        this.max_retries = source["max_retries"];
	        this.queue_id = source["queue_id"];
	        this.priority = source["priority"];
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
	
	export class QueueModel {
	    id: number;
	    name: string;
	    status: number;
	    max_concurrent: number;
	    speed_limit: number;
	    auto_start: boolean;
	    schedule_enabled: boolean;
	    // Go type: time
	    schedule_start?: any;
	    // Go type: time
	    schedule_stop?: any;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    download_order: number[];
	
	    static createFrom(source: any = {}) {
	        return new QueueModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.max_concurrent = source["max_concurrent"];
	        this.speed_limit = source["speed_limit"];
	        this.auto_start = source["auto_start"];
	        this.schedule_enabled = source["schedule_enabled"];
	        this.schedule_start = this.convertValues(source["schedule_start"], null);
	        this.schedule_stop = this.convertValues(source["schedule_stop"], null);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.download_order = source["download_order"];
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


const API_BASE = "http://localhost:8081";

interface ApiOpts {
    method?: string;
    body?: any;
    auth?: boolean;
}

async function apiCall<T = any>(path: string, opts: ApiOpts = {}): Promise<T> {
    const headers: Record<string, string> = {
        "Content-Type": "application/json",
    };

    if (opts.auth !== false) {
        const token = userToken();
        if (token) {
            headers["Authorization"] = "Bearer " + token;
        }
    }

    const res = await fetch(API_BASE + path, {
        method: opts.method || "GET",
        headers: headers,
        body: opts.body ? JSON.stringify(opts.body) : undefined,
    });

    if (res.status === 401) {
        const hadToken = userToken().length > 0;
        clearUser();
        if (hadToken && window.location.pathname !== "/login") {
            window.location.assign("/login");
        }
        throw new Error("unauthorized");
    }

    const text = await res.text();
    let data: any = null;
    if (text) {
        try {
            data = JSON.parse(text);
        } catch (_) {
            data = text;
        }
    }

    if (!res.ok) {
        const msg = (data && data.error) ? data.error : "HTTP " + res.status;
        throw new Error(msg);
    }

    return data as T;
}

interface Container {
    id: string;
    name: string;
    image: string;
    status: string;
}

interface ContainerStats {
    id: string;
    cpu_percent: number;
    memory_mb: number;
    memory_limit_mb: number;
}

interface LoginResp {
    token: string;
    email: string;
}

async function apiLogin(email: string, password: string): Promise<LoginResp> {
    const resp = await apiCall<LoginResp>("/login", {
        method: "POST",
        body: { email: email, password: password },
        auth: false,
    });
    saveUser({ id: resp.email, email: resp.email, token: resp.token });
    return resp;
}

async function apiLogout(): Promise<void> {
    try {
        await apiCall("/logout", { method: "POST" });
    } catch (_) {
        // ignore
    }
    clearUser();
}

function apiListContainers(): Promise<Container[]> {
    return apiCall<Container[]>("/containers");
}

function apiStartContainer(id: string): Promise<any> {
    return apiCall("/containers/" + id + "/start", { method: "POST" });
}

function apiStopContainer(id: string): Promise<any> {
    return apiCall("/containers/" + id + "/stop", { method: "POST" });
}

function apiRestartContainer(id: string): Promise<any> {
    return apiCall("/containers/" + id + "/restart", { method: "POST" });
}

function apiContainerLogs(id: string): Promise<{ id: string; logs: string }> {
    return apiCall("/containers/" + id + "/logs");
}

function apiContainerStats(id: string): Promise<ContainerStats> {
    return apiCall("/containers/" + id + "/stats");
}

function apiContainerInspect(id: string): Promise<any> {
    return apiCall("/containers/" + id + "/inspect");
}

function pierrotUpdate(): void {
    const fn = (window as any).__pierrotUpdate;
    if (typeof fn === "function") {
        fn();
    }
}

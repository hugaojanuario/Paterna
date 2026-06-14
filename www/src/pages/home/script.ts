async function homeFetchContainerData(id: string) {
    const result: any = {
        name: "",
        image: "",
        status: "",
        logs: "",
        cpu: 0,
        mem: 0,
    };

    if (!id) {
        return result;
    }

    try {
        const list = await apiListContainers();
        if (list) {
            const found = list.find(function (c: any) { return c.id === id; });
            if (found) {
                result.name = found.name;
                result.image = found.image;
                result.status = found.status;
            }
        }
    } catch (_) { /* ignore */ }

    try {
        const logResp = await apiContainerLogs(id);
        result.logs = logResp.logs || "";
    } catch (_) { /* ignore */ }

    try {
        const statsResp = await apiContainerStats(id);
        result.cpu = Math.round((statsResp.cpu_percent || 0) * 100) / 100;
        result.mem = Math.round((statsResp.memory_mb || 0) * 10) / 10;
    } catch (_) { /* ignore */ }

    return result;
}

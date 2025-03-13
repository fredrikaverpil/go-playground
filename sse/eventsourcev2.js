const eventSource = new EventSource("http://127.0.0.1:8080/events");
const mem = document.getElementById("mem");
const cpu = document.getElementById("cpu");

eventSource.addEventListener("mem", (event) => {
	const memData = JSON.parse(event.data);
	mem.innerHTML = `Total: ${memData.total}<br>` +
		`Free: ${memData.free}<br>` +
		`Available: ${memData.available}<br>` +
		`Used: ${memData.used}<br>` +
		`Used %: ${memData.usedPercent.toFixed(2)}%`;
});

eventSource.addEventListener("cpu", (event) => {
	const cpuData = JSON.parse(event.data);
	cpu.innerHTML = `User: ${cpuData.user.toFixed(2)}<br>` +
		`System: ${cpuData.system.toFixed(2)}<br>` +
		`Idle: ${cpuData.idle.toFixed(2)}`;
});

eventSource.onerror = (error) => {
	console.error("EventSource failed:", error);
}

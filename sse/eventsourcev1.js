const eventSource = new EventSource("http://127.0.0.1:8080/events");
const mem = document.getElementById("mem");
const cpu = document.getElementById("cpu");

eventSource.addEventListener("mem", (event) => {
	mem.textContent = event.data;
});

eventSource.addEventListener("cpu", (event) => {
	cpu.textContent = event.data;
});

eventSource.onerror = (error) => {
	console.error("EventSource failed:", error);
}

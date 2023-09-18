document.addEventListener('DOMContentLoaded', async () => {
  const urlParams = new URLSearchParams(window.location.search);
  const sessionId = urlParams.get("session_id")

  if (sessionId) {
    document.querySelector("#sessionId").value = sessionId;

    const session = await fetch(`http://localhost:4242/checkout-session?sessionId=${sessionId}`).then(r => r.json());
    var sessionJSON = JSON.stringify(session, null, 2);
    document.querySelector("pre").textContent = sessionJSON;
  }
});

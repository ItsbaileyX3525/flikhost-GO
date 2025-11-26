const form = document.getElementById("image-form")
let turnstileToken = ""

form.addEventListener("submit", async (e) => {
    e.preventDefault()

    if (!turnstileToken) {
        console.log("Waiting for turnstile verif")
        return
    }

    const formData = new FormData(form)
    formData.append("token", turnstileToken)

    const resp = await fetch("/api/uploadImage", {
        method: "POST",
        body: formData
    })

    if (!resp.ok) {
        console.log("Something went wrong")
        return
    }

    const data = await resp.json()

    console.log(data)
    console.log(data.message)
})

  function onTurnstileSuccess(token) {
    console.log("Turnstile success:", token);
    turnstileToken = token
  }
  function onTurnstileError(errorCode) {
    console.error("Turnstile error:", errorCode);
  }
  function onTurnstileExpired() {
    console.warn("Turnstile token expired");
  }
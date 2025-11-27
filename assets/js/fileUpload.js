const form = document.getElementById("upload-form")
let turnstileToken = ""
let widgetID = null

form.addEventListener("submit", async (e) => {
	e.preventDefault()

	let URL = "/api/uploadImage"

	const input = document.getElementById("imageUpload")
	if (input && input.getAttribute("accept") !== "image/png") {
		URL = "/api/uploadFile"
	}

	if (!turnstileToken) {
		console.log("Waiting for turnstile verification")
		return
	}

	const formData = new FormData(form)
	formData.append("token", turnstileToken)

	const resp = await fetch(URL, {
		method: "POST",
		body: formData
	})

	if (!resp.ok) {
		turnstileToken = ""
		turnstile.reset(widgetID)
		return
	}

	const data = await resp.json()
	console.log(data)
	console.log(data.message)

	turnstileToken = ""
	turnstile.reset(widgetID)
})

window.onloadTurnstileCallback = function () {
	widgetID = turnstile.render("#turnstile-container", {
		sitekey: "0x4AAAAAACDFfOKm7uvwfqiR",
		callback: function (token) {
			turnstileToken = token
		}
	})

}

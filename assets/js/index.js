const form = document.getElementById("image-form")

form.addEventListener("submit", async (e) => {
    e.preventDefault()

    const formData = new FormData(form)

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
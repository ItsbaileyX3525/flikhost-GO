const mlgAudio = document.getElementById("mlg")

let keystrokes = []
const perfectCell = [
    "https://flik.host/images/3ougCg.gif",
    "https://flik.host/images/2OOTCL.gif"
]

function mlgMode() {
    const mlgGif = document.createElement("img")
    mlgGif.src = "https://flik.host/images/QrURN.gif"
    mlgGif.id = "mlgGif"
    document.body.appendChild(mlgGif)
    mlgAudio.volume = .1
    mlgAudio.play()
    setTimeout(() => {
        mlgAudio.paused = true
        mlgGif.remove()
    }, 20000);
}

function kawaiiMode() {
    console.log("kawaii mode")
    document.body.style.backgroundImage = `url("${perfectCell[Math.floor(Math.random() * 2)]}")`;
    document.body.style.backgroundColor = '';
}

function checkCombo() {
    switch (keystrokes.join("")) {
        case "mlg":
            mlgMode()
            break
        case ":3":
            kawaiiMode()
            break
    }
}

document.addEventListener("keypress", (e) => {
    if (e.code === "Enter") {
        checkCombo()
        keystrokes = []
        return
    }
    keystrokes.push(e.key)
    console.log(keystrokes)
})
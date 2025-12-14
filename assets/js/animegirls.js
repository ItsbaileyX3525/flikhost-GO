const animegirlImage = document.querySelector(".anime-girl-image")
const animegirlVideo = document.getElementById("bg")
const clickText = document.getElementById("click")

document.addEventListener("DOMContentLoaded", () => {
    if (animegirlImage) {
        animegirlImage.addEventListener("click", () => {
            window.location.href = "/animegirls.html"
        })
        animegirlImage.src = `/assets/images/animegirls/animegirl${Math.floor(Math.random() * 3) + 1}.png`
    }
    if (animegirlVideo) {
        animegirlVideo.addEventListener("click", () => {
            animegirlVideo.play()
            animegirlVideo.volume = .3
            if (clickText) {
                clickText.style.display = "none"
            }
        })

        animegirlVideo.addEventListener("ended", () => {
            window.location.href = "/"
        })
    }
})
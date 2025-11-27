document.addEventListener('DOMContentLoaded', function () {
    const enterSiteBtn = document.getElementById('enterSiteBtn');
    const overlay = document.getElementById('overlay');
    const infoDiv = document.getElementById('infoDiv');
    const cookedAudio = document.getElementById('cookedAudio');

    enterSiteBtn.addEventListener('click', function () {
        enterSiteBtn.disabled = true;

        setTimeout(function () {
            overlay.style.opacity = '0';
            overlay.style.pointerEvents = 'none';
            overlay.style.transition = 'opacity 1.5s ease-out';
        }, 1000);
        
        infoDiv.classList.remove('hidden');

        cookedAudio.play();

        var url = "/api/info"
        if (window.location.hostname === "localhost") {
            url = "/api/info?testIP=142.250.117.100"
        }

        setTimeout(function () {
            fetch(url)
                .then(response => response.json())
                .then(data => {
                    let values = Object.values(data);
                    let index = 0;

                    function displayNext() {
                        if (index < values.length) {
                            let item = document.createElement('h1');
                            item.className = 'info-item';
                            item.textContent = values[index];
                            infoDiv.appendChild(item);
                            index++;
                            setTimeout(displayNext, 800);
                        }
                    }

                    displayNext();
                })
                .catch(error => console.error('Error fetching info:', error));
        }, 1000);
    });
});

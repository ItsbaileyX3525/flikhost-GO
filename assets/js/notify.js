function notify(message, type = "info", duration = 2200) {
    const notifContainer = document.getElementById("notif-container");
    const notif = document.createElement("div");


    notif.classList.add("notification", type);
    notif.innerText = message;


    notifContainer.appendChild(notif);

    setTimeout(() => {
        notif.style.animation = "fadeOut 0.3s ease forwards";
        setTimeout(() => notif.remove(), 300);
        
    }, duration);
}

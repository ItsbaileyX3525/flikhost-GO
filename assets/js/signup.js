let widgetId = null;
let turnstileToken = '';
const signupForm = document.getElementById('signupForm');

function initTurnstile() {
    if (!window.turnstile) {
        setTimeout(initTurnstile, 500);
        return;
    }
    
    var container = document.getElementById('turnstile-container');
    if (container) {
        widgetId = turnstile.render('#turnstile-container', {
            sitekey: '0x4AAAAAACDFfOKm7uvwfqiR',
            theme: 'dark',
            callback: function(token) {
                turnstileToken = token;
            }
        });
    }
}

function resetTurnstile() {
    if (window.turnstile && widgetId !== null) {
        turnstile.reset(widgetId);
        turnstileToken = '';
    }
}

document.addEventListener('DOMContentLoaded', function() {
    initTurnstile();
});

signupForm.addEventListener('submit', async function(event) {
    event.preventDefault();

    if (!turnstileToken) {
        notify('Please verify you are a human', 'error');
        return;
    }

    var tosCheckbox = document.getElementById('TOSCheckbox');
    if (!tosCheckbox.checked) {
        notify('Please agree to the ToS', 'error');
        return;
    }

    var username = document.getElementById('username').value;
    var email = document.getElementById('email').value;
    var password = document.getElementById('password').value;

    console.log(username,password,email)

    const resp = await fetch("/api/createAccount", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "username" : username,
            "password" : password,
            "email" : email,
            "turnstile" : turnstileToken
        })
    })

    if (!resp.ok){
        notify("Fetch request failed")
        return
    }

    const data = await resp.json()

    if (data.status == "success") {
        notify(data.message, "success")
        setTimeout(() => {
            window.location.href = "/"
        }, 2000);
    } else {
        notify(data.message, "error")
    }

    resetTurnstile();
});
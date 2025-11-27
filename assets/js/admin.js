document.getElementById('adminForm').addEventListener('submit', function (e) {
    e.preventDefault();

    const contentDiv = document.getElementById('content');
    
    setTimeout(function () {
        contentDiv.innerHTML = `
            <br><br><br>
            <h1>this is NOT the admin page</h1><br>
            <img src='/assets/images/ryan-gosling-angry.gif' alt='gif of ryan gosling.. crashing out'>
            <!-- AND YOU ARE STILL LOOKING AT THE SOURCE HAHAHAH, NOTHIN' HERE -->
        `;
    }, 400);
});

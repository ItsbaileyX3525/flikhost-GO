let validationKey = "";

const formatter = new Intl.ListFormat('en', {
    style: 'long',
    type: 'conjunction',
});
function rainbow(string) {
return lolcat.rainbow(function(char, color) {
    char = $.terminal.escape_brackets(char);
    return `[[;${hex(color)};]${char}]`;
}, string).join('\n');
}

function hex(color) {
return '#' + [color.red, color.green, color.blue].map(n => {
    return n.toString(16).padStart(2, '0');
}).join('');
}
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

async function remove_image_from_db(imageID){
    const data = {
        image_id: imageID,
        server: validationKey
    };

    try {
        const response = await $.ajax({
            url: '/api/deleteImage',
            type: 'POST',
            data: data,
            dataType: 'json'
        });

        return { success: response.success, ban_response: response.message };

    } catch (error) {
        console.error("AJAX error:", error);
        return { success: false, ban_response: "Error removing image." };
    }
}

async function getValidationKey(password) {
    data = {
        gimmieServerKey: password
    }

    try {
        const response = await $.ajax({
            url: '/api/validateKey',
            type: 'POST',
            data: data,
            dataType: 'json'
        });

        if (response.success) {
            return { success: response.success, vaildationkey: response.key };
        }else{
            return { success: response.success, error: response.message };
        }

    } catch (error) {
        console.error("AJAX error:", error);
        return { success: false, ban_response: "Error getting validation key." };
    }
}

const commands = {
    async login(password){
        term.echo(`[[;${hex({ red: 255, green: 165, blue: 0 })};]Attempting to login...]`);
        
        const ban_result = await getValidationKey(password);

        await sleep(1000);

        if (ban_result.success) {
            response = `[[;${hex({ red: 124, green: 252, blue: 0 })};]Retrieved validation key: ]`;
            validationKey = ban_result.vaildationkey;
        } else {
            response = `[[;${hex({ red: 255, green: 0, blue: 0 })};]Failed to get vaildation key. Error: ${ban_result.error} Current key: ]`;
        }

        term.echo(response + validationKey);
    },
    help() {
        term.echo(`List of available commands: ${help}`);
    },
    echo(args){
        term.echo(`[[;${hex({ red: 255, green: 165, blue: 0 })};]${args}]`);
    },
    check_key(){
        term.echo(`[[;${hex({ red: 255, green: 165, blue: 0 })};]Current validation key: ]${validationKey}`);
    },
    async cowsay(args){
        const apiUrl = "https://cowsay.morecode.org/say"; 
        const format = "json";
        await fetch(`/api/proxy?url=${encodeURIComponent(apiUrl)}&message=${encodeURIComponent(args)}&format=${encodeURIComponent(format)}`)        
        .then(response => response.text())
        .then(text => {
          let cowsay_response = JSON.parse(text);
          term.echo(cowsay_response.cow);

        })
        .then(data => console.log(data))
        .catch(error => console.error("Error fetching data:", error));
    },
    async kanye(){
        await fetch("https://api.kanye.rest/")
        .then(response => response.json())
        .then(data => term.echo(data.quote))
        .catch(error => console.error("Error fetching data:", error));
      
    },
    lolcat(args){
        const message = rainbow(args);
        term.echo(message);
    },
    async figlet(args){
        try {
            const fontUrl = '/assets/fonts/big.flf';
            const fontResponse = await fetch(fontUrl);
            if (!fontResponse.ok) {
                term.echo('Error loading font');
                return;
            }
            const fontData = await fontResponse.text();
            
            figlet.parseFont('big', fontData);
            figlet.text(args, { font: 'big' }, function(err, data) {
                if (err) {
                    console.log('Figlet error, falling back to plain text');
                    term.echo(args);
                    return;
                }
                term.echo(data);
            });
        } catch (error) {
            console.error('Figlet error:', error);
            term.echo(args);
        }
    },
    async cat_fact(){
        await fetch("https://meowfacts.herokuapp.com/")
        .then(response => response.json())
        .then(data => term.echo(data.data[0]))
        .catch(error => console.error("Error fetching data:", error));
      
    },
    async remove_image(imageID){
        term.echo(`[[;${hex({ red: 255, green: 165, blue: 0 })};]Attempting to remove image, ${imageID}...]`);
        
        const remove_result = await remove_image_from_db(imageID);

        await sleep(1000);

        if (remove_result.success) {
            response = `[[;${hex({ red: 124, green: 252, blue: 0 })};]removed image, ${imageID} from database!]`;
        } else {
            response = `[[;${hex({ red: 255, green: 0, blue: 0 })};]Failed to remove image. Error: ${remove_result.ban_response}]`;
        }

        term.echo(response);
    }
};


const command_list = Object.keys(commands);
const help = formatter.format(command_list);

const term = $('body').terminal(commands);


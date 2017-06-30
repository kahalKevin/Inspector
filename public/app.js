new Vue({
    el: '#app',
    data: {
        ws: null, // Our websocket
        pythonscript: '', //python script to be uploaded
        companionfile : '', //companion file of py script
        newMsg: '', // Holds new messages to be sent to the server
        own_address: '', // Holds our own address
        to_address: '', //Where will the chat sent
        chatContent: {}, // A running list of chat messages displayed on the screen
        chatContentOne: '',
        contact_list: [],
        title_assertion: null,
        duration: null,
        interval: null,
        email: null, // Email address used for grabbing an avatar
        username: null, // Our username
        joined: false // True if email and username have been filled in
    },

    created: function() {
        var self = this;
        var own_address;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            console.log(msg);
    
        });
    },

    methods: {
        send: function () {
            if (this.newMsg != ''  &&  this.to_address != '') {
                this.ws.send(
                    JSON.stringify({
                        to_address: this.to_address,
                        from: this.own_address,
                        type : 'message',
                        email: this.email,
                        username: this.username,
                        message: $('<p>').html(this.newMsg).text() // Strip out html
                    }
                ));
                // Add this to our chat window too
                this.chatContent[this.to_address] += '<div class="chip">'
                        + this.username
                    + '</div>'
                    + emojione.toImage( $('<p>').html(this.newMsg).text() ) + '<br/>'; // Parse emojis

                var element = document.getElementById('chat-messages');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom

                this.newMsg = ''; // Reset newMsg
                this.to_address = ''; // Reser to_address
            }
        },

        selectContact : function (selected){
            this.chatContentOne = this.chatContent[selected];
            this.to_address = selected;
        },

        filesChange: function(fileUpload) {
            this.pythonscript = fileUpload[0];
        },

        filesChange2: function(fileUpload) {
            this.companionfile = fileUpload[0];
        },        

        startAssertion: function() {
            this.ajaxRequest = true;
            var startAssertionURL = 'http://' + window.location.host + '/start'
            var formData = new FormData();
            formData.append('title', this.title_assertion);
            formData.append('duration', this.duration);
            formData.append('interval', this.interval);
            formData.append('pythonscript', this.pythonscript);
            formData.append('companionfile', this.companionfile);
            this.$http.post(startAssertionURL, formData, {
               headers: {
                   'Content-Type': 'multipart/form-data'
               }
            }).then(response => {
               console.log(response);
            }, response => {});

            this.title_assertion = null;
            this.duration = null;
            this.interval = null;
            this.pythonscript = null;
            this.companionfile = null;
            document.getElementById("fileInput1").value = "";
            document.getElementById("fileInput2").value = "";
        }

    }
});
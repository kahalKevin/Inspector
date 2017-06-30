new Vue({
    el: '#app',
    data: {
        ws: null, // Our amazing websocket
        pythonscript: '', //python script to be uploaded
        companionfile : '', //companion file of py script
        monitoringDataContent: '',
        title_assertion: null,
        duration: null,
        interval: null
    },

    created: function() {
        var self = this;
        var own_address;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            var content = '';
            for (var key in msg) {
              if (msg.hasOwnProperty(key)) {
                content += 'Assertion Title: <font size="4"><b>' + emojione.toImage( $('<p>').html(msg[key][0].title).text() ) + '</b></font>';
                content += ' | id: <b>' + emojione.toImage( $('<p>').html(msg[key][0].pyscript).text() ) + '</b>';
                content += ' | timelog: <b>' + emojione.toImage( $('<p>').html(msg[key][0].timelog).text() ) + '</b>';
                if (msg[key].length > 1) {
                    content += '<ul>'
                    for (var i = 1; i < msg[key].length; i++) {
                       content += '<li>'
                       var executed = ''
                       if (msg[key][i].times == -1){
                        executed = 'Stopped 👎'
                       }else{
                        executed = msg[key][i].times
                       }
                       content += 'Execution number: ' + emojione.toImage( $('<p>').html(executed).text() );
                       if (msg[key][i].cleared) {
                        content += ' | Result: <font color="green">' + emojione.toImage( $('<p>').html('Cleared 😄').text() ) + '</font>';
                       }else{
                        content += ' | Result: <font color="red">' + emojione.toImage( $('<p>').html('Fail 😓').text() ) + '</font>';
                       }
                       content += ' | Timelog: <b>' + emojione.toImage( $('<p>').html(msg[key][i].timelog).text() ) + '</b>';
                       if (msg[key][i].entity != null || msg[key][i].envent != null){
                        content += '<div class="row">'
                        if (msg[key][i].entity != null){
                            content += '<div class="col s6">'
                            content += '<ul>'
                            for (var j = 0; j < msg[key][i].entity.length; j++) {
                                content += '<li>'
                                content += 'Validation on: <b>' + emojione.toImage( $('<p>').html(msg[key][i].entity[j].key).text() ) + '</b>';
                                if (msg[key][i].entity[j].value) {
                                    content += ' | Result: <font color="green">' + emojione.toImage( $('<p>').html('Passed').text() ) + '</font>';
                                }else{
                                    content += ' | Result: <font color="red">' + emojione.toImage( $('<p>').html('Fail').text() ) + '</font>';   
                                }
                                content += ' | Added info: <b>' + emojione.toImage( $('<p>').html(msg[key][i].entity[j].info).text() ) + '</b>';
                                content += '</li>'
                            }
                            content += '</ul>'
                            content += '</div>'
                        }
                        if (msg[key][i].envent != null){
                            content += '<div class="col s6">'
                            content += '<ul>'
                            for (var j = 0; j < msg[key][i].envent.length; j++) {
                                content += '<li>'
                                content += 'Env Name: ' + emojione.toImage( $('<p>').html(msg[key][i].envent[j].env).text() );
                                content += ' | Free memory: <b>' + emojione.toImage( $('<p>').html(msg[key][i].envent[j].free).text() ) + '</b>';
                                content += ' | Added info: <b>' + emojione.toImage( $('<p>').html(msg[key][i].envent[j].info).text() ) + '</b>';
                                content += '</li>'
                            }
                            content += '</ul>'
                            content += '</div>'
                        }
                        content += '</div>'
                       }
                       content += '</li>'
                    }
                    content += '</ul>'
                }
                content += '<hr/>'
              }
            }
            self.monitoringDataContent = content;
        });
    },

    methods: {
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
            });
            // compatibility problem with Safari
            // .then(response => {
            //    console.log(response);
            // }, response => {});

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
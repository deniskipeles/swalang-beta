const path = require('path');
const { LanguageClient } = require('vscode-languageclient/node');

let client;

function activate(context) {
    console.log('Starting Swalang LSP...');

    // Point this to your compiled binary!
    // Since vscode-ext is inside swalang-beta, we go up one directory to builds/
    const serverPath = path.join(__dirname, '..', 'builds', 'swalang-lsp'); 
    // On Windows use: 'swalang-lsp.exe'

    const serverOptions = {
        command: serverPath,
        args: []
    };

    const clientOptions = {
        // Trigger this LSP on our custom language ID
        documentSelector: [{ scheme: 'file', language: 'swalang' }]
    };

    client = new LanguageClient(
        'swalangLSP',
        'Swalang Language Server',
        serverOptions,
        clientOptions
    );

    client.start();
}

function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

module.exports = {
    activate,
    deactivate
};
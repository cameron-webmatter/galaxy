import * as path from 'path';
import * as vscode from 'vscode';
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration('gxc');
  
  if (!config.get('lsp.enable')) {
    return;
  }

  const serverPath = config.get<string>('lsp.serverPath') || 'gastro';
  
  const serverOptions: ServerOptions = {
    command: serverPath,
    args: ['lsp-server', '--stdio'],
    transport: TransportKind.stdio
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: 'file', language: 'gxc' }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher('**/*.gxc')
    }
  };

  client = new LanguageClient(
    'gxcLanguageServer',
    'GXC Language Server',
    serverOptions,
    clientOptions
  );

  client.start();
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}

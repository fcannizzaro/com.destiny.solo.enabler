import { Action, AppearDisappearEvent, BaseAction } from '@stream-deck-for-node/sdk';
import { sd } from './index';
import { dirname, join } from 'path';
import { WebSocket } from 'ws';
import { exec } from 'child_process';

const delay = (ms: number) => new Promise((res) => setTimeout(res, ms));

const getBasePath = (path: string) =>
  process['pkg'] ? join(dirname(process.execPath), path) : join(process.cwd(), './plugin/', path);

const Elevate = getBasePath('firewall/elevate.cmd');
const FirewallExe = getBasePath('firewall/firewall-changer.exe');
const PermissionIcon = getBasePath('icons/solo-mode-permission.png');

@Action('toggle')
export class Toggle extends BaseAction {
  // flag to init the firewall changer executable
  firewallChangerStarted = false;

  // ws to the firewall changer
  ws: WebSocket;

  initWS(res: () => void) {
    try {
      this.ws = new WebSocket('ws://localhost:33334');
      this.ws.on('message', (data) => {
        this.contexts.forEach((context) => {
          sd.setImage(context);
          sd.setState(context, data.toString() === 'true' ? 1 : 0);
        });
      });
      this.ws.on('open', () => {
        this.ws.send('info');
        res();
      });
    } catch (e) {
      setTimeout(() => this.initWS(res), 250);
    }
  }

  startFirewallChanger() {
    return new Promise<void>((res) => {
      exec(`${Elevate} ${FirewallExe}`, async (err) => {
        if (!err) {
          await delay(250);
          this.firewallChangerStarted = true;
          this.initWS(res);
          const onExit = () => this.ws?.send('exit');
          process.on('exit', onExit);
          process.on('SIGINT', onExit);
          process.on('SIGUSR1', onExit);
          process.on('SIGUSR2', onExit);
          process.on('uncaughtException', onExit);
        }
      });
    });
  }

  async onAppear(e: AppearDisappearEvent) {
    if (!this.firewallChangerStarted) {
      sd.setImage(e.context, PermissionIcon);
    }
  }

  async onSingleTap() {
    if (!this.firewallChangerStarted) {
      return this.startFirewallChanger();
    }
    this.ws?.send('toggle');
  }
}

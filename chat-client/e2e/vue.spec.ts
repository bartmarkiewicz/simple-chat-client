import { test, expect } from '@playwright/test';
import { spawn, ChildProcess } from 'node:child_process';
import net from 'node:net';

// Wait for Go web server to start-up
async function waitForPort(port: number, host = '127.0.0.1', timeoutMs = 10000): Promise<void> {
  const start = Date.now();
  return new Promise((resolve, reject) => {
    const tryOnce = () => {
      const socket = net.createConnection({ port, host });
      socket.once('connect', () => {
        socket.destroy();
        resolve();
      });
      socket.once('error', () => {
        socket.destroy();
        if (Date.now() - start > timeoutMs) {
          reject(new Error(`Timeout waiting for ${host}:${port}`));
        } else {
          setTimeout(tryOnce, 100);
        }
      });
    };
    tryOnce();
  });
}

let wsServerProc: ChildProcess | undefined;

test.beforeAll(async () => {
  // Start the Go websocket server
  wsServerProc = spawn('go', ['run', 'main.go', 'server.go'], {
    cwd: '..',
    stdio: 'inherit',
    shell: false,
  });
  await waitForPort(8080);
});

test.afterAll(async () => {
  if (wsServerProc && !wsServerProc.killed) {
    wsServerProc.kill();
  }
});

test('websocket chat relays messages between two clients', async ({ browser }) => {
  const context1 = await browser.newContext();
  const page1 = await context1.newPage();
  await page1.goto('/');
  await expect(page1.locator('form.form-input')).toBeVisible();

  const context2 = await browser.newContext();
  const page2 = await context2.newPage();
  await page2.goto('/');
  await expect(page2.locator('form.form-input')).toBeVisible();

  await page1.getByPlaceholder('Type a message...').fill('Hello from P1');
  await page1.getByRole('button', { name: 'Send' }).click();

  const msgOnP2 = page2.locator('li', { hasText: 'Hello from P1' });
  await expect(msgOnP2).toHaveCount(1);
  await expect(msgOnP2.locator('div.system')).toHaveCount(0);
  const sender2 = await msgOnP2.locator('strong').innerText();
  expect(sender2).not.toContain('System');

  const p1OwnCount = await page1.locator('li', { hasText: 'Hello from P1' }).count();
  expect(p1OwnCount).toBe(1);

  await page2.getByPlaceholder('Type a message...').fill('Reply from P2');
  await page2.getByRole('button', { name: 'Send' }).click();

  const msgOnP1 = page1.locator('li', { hasText: 'Reply from P2' });
  await expect(msgOnP1).toHaveCount(1);
  await expect(msgOnP1.locator('div.system')).toHaveCount(0);
  const sender1 = await msgOnP1.locator('strong').innerText();
  expect(sender1).not.toContain('System');

  const p2OwnCount = await page2.locator('li', { hasText: 'Reply from P2' }).count();
  expect(p2OwnCount).toBe(1);

  await context1.close();
  await context2.close();
});

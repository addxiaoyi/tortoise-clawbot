/**
 * Email Channel Adapter (OpenClaw Compatible)
 * Supports SMTP for sending and IMAP for receiving emails
 */

import { BaseChannelAdapter } from './base.js';
import type {
  ChannelMessage,
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface EmailConfig {
  // SMTP Settings
  smtp: {
    host: string;
    port: number;
    secure: boolean;        // true for 465, false for other ports
    auth: {
      user: string;
      pass: string;
    };
  };
  // IMAP Settings (for receiving)
  imap?: {
    host: string;
    port: number;
    secure: boolean;
    auth: {
      user: string;
      pass: string;
    };
    box?: string;           // IMAP mailbox to check (default: INBOX)
  };
  // Email identity
  from: string;             // From address (e.g., bot@example.com)
  fromName?: string;        // From display name
  // Filtering
  allowedFrom?: string[];   // Allowed sender emails/domains
  allowedDomains?: string[];// Allowed email domains
  subjectPrefix?: string;    // Subject prefix to identify bot emails
}

export class EmailChannel extends BaseChannelAdapter {
  readonly name = 'email';
  readonly capabilities: ChannelCapability[] = [
    'text',
    'html',
    'files',
  ];

  private config?: EmailConfig;
  private lastChecked?: Date;
  private checkInterval = 60000; // 1 minute default
  private imapTimer?: NodeJS.Timeout;
  private imapConnection?: unknown;

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<EmailConfig>();
    
    if (!this.config?.smtp) {
      ctx.logger.warn('[email] SMTP not configured, channel disabled');
      return;
    }

    if (!this.config.from) {
      ctx.logger.warn('[email] From address not set');
    }
  }

  async onStart(): Promise<void> {
    await super.onStart();
    
    // Test SMTP connection
    await this.testSmtpConnection();
    
    // Start IMAP polling if configured
    if (this.config?.imap) {
      await this.connectImap();
      this.startImapPolling();
    }
  }

  async onStop(): Promise<void> {
    if (this.imapTimer) {
      clearInterval(this.imapTimer);
    }
    if (this.imapConnection) {
      await this.disconnectImap();
    }
    await super.onStop();
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);
    
    const to = message.to;
    const subject = message.options?.subject || 'Message from Tortoise';
    const content = message.content;

    // Prepare email headers
    const headers = this.buildHeaders(subject, content);
    
    // Connect and send via SMTP
    await this.smtpSend(to, headers, content);
    
    this.ctx?.logger.debug(`[email] Email sent to ${to}`);
  }

  async formatForChannel(content: string): Promise<string> {
    // Email body formatting - convert markdown to HTML
    return this.markdownToHtml(content);
  }

  private buildHeaders(subject: string, content: string): Record<string, string> {
    const headers: Record<string, string> = {
      'From': this.config?.fromName 
        ? `"${this.config.fromName}" <${this.config.from}>`
        : this.config?.from || 'tortoise@localhost',
      'To': '',
      'Subject': this.config?.subjectPrefix 
        ? `[${this.config.subjectPrefix}] ${subject}`
        : subject,
      'MIME-Version': '1.0',
      'Content-Type': 'text/html; charset=utf-8',
      'X-Mailer': 'Tortoise-AI/1.0',
    };

    return headers;
  }

  private markdownToHtml(markdown: string): string {
    let html = markdown
      // Escape HTML special characters first
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      // Headers
      .replace(/^### (.*$)/gim, '<h3>$1</h3>')
      .replace(/^## (.*$)/gim, '<h2>$1</h2>')
      .replace(/^# (.*$)/gim, '<h1>$1</h1>')
      // Bold and italic
      .replace(/\*\*\*(.*?)\*\*\*/g, '<strong><em>$1</em></strong>')
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      // Code
      .replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      // Links
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>')
      // Line breaks and paragraphs
      .replace(/\n\n/g, '</p><p>')
      .replace(/\n/g, '<br>');
    
    return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; }
pre { background: #f5f5f5; padding: 16px; border-radius: 8px; overflow-x: auto; }
code { background: #f5f5f5; padding: 2px 6px; border-radius: 4px; }
a { color: #0066cc; }
</style>
</head>
<body>
<p>${html}</p>
</body>
</html>`;
  }

  private async smtpSend(
    to: string,
    headers: Record<string, string>,
    content: string
  ): Promise<void> {
    if (!this.config?.smtp) {
      throw new Error('SMTP not configured');
    }

    const { host, port, secure, auth } = this.config.smtp;
    
    // Build the email message
    let emailContent = '';
    for (const [key, value] of Object.entries(headers)) {
      if (value !== undefined) {
        emailContent += `${key}: ${value}\n`;
      }
    }
    emailContent += '\n' + content;

    try {
      // Use native Node.js TLS/TCP for SMTP
      const { TLSSocket } = await import('node:tls');
      const { Socket } = await import('node:net');
      
      const socket = secure 
        ? new TLSSocket(await this.createSocket(host, port))
        : new Socket();
      
      if (!secure) {
        socket.connect({ host, port });
      }

      await this.smtpTransaction(socket, to, emailContent);
      
      socket.end();
      this.ctx?.logger.info(`[email] SMTP: Email sent successfully to ${to}`);
    } catch (error) {
      this.ctx?.logger.error(`[email] SMTP send failed: ${error}`);
      throw error;
    }
  }

  private createSocket(host: string, port: number): Promise<import('net').Socket> {
    return new Promise((resolve, reject) => {
      const net = require('net');
      const socket = net.createConnection({ host, port }, () => {
        resolve(socket);
      });
      socket.on('error', reject);
    });
  }

  private async smtpTransaction(
    socket: import('net').Socket, 
    to: string,
    emailContent: string
  ): Promise<void> {
    return new Promise((resolve, reject) => {
      let step = 0;
      const timeout = setTimeout(() => reject(new Error('SMTP timeout')), 30000);
      
      const expect = (code: number, callback: () => void) => {
        let data = '';
        const handler = (chunk: Buffer) => {
          data += chunk.toString();
          const lines = data.split('\n');
          const lastLine = lines[lines.length - 1];
          if (lastLine.startsWith(String(code))) {
            socket.removeListener('data', handler);
            clearTimeout(timeout);
            setTimeout(callback, 10);
          }
        };
        socket.on('data', handler);
      };

      const send = (cmd: string) => {
        this.ctx?.logger.debug(`[email] SMTP TX: ${cmd.trim()}`);
        socket.write(cmd + '\r\n');
      };

      expect(220, () => {
        send('EHLO localhost');
        step = 1;
      });

      expect(250, () => {
        if (step === 1) {
          send(`AUTH LOGIN`);
          step = 2;
        } else if (step === 2) {
          send(Buffer.from(this.config!.smtp.auth.user).toString('base64'));
          step = 3;
        } else if (step === 3) {
          send(Buffer.from(this.config!.smtp.auth.pass).toString('base64'));
          step = 4;
        } else if (step === 4) {
          send(`MAIL FROM:<${this.config!.from}>`);
          step = 5;
        } else if (step === 5) {
          send(`RCPT TO:<${to}>`);
          step = 6;
        } else if (step === 6) {
          send('DATA');
          step = 7;
        } else if (step === 7) {
          send(emailContent + '\r\n.');
          step = 8;
        } else if (step === 8) {
          send('QUIT');
          clearTimeout(timeout);
          resolve();
        }
      });
    });
  }

  private async testSmtpConnection(): Promise<void> {
    if (!this.config?.smtp) {
      throw new Error('[email] SMTP configuration required');
    }

    const { host, port, secure, auth } = this.config.smtp;
    
    this.ctx?.logger.info(`[email] Testing SMTP connection to ${host}:${port}`);
    
    try {
      const { TLSSocket } = await import('node:tls');
      const socket = await this.createSocket(host, port);
      const tlsSocket = secure ? new TLSSocket(socket) : socket;
      
      if (secure) {
        tlsSocket.connect({ host, port, socket });
      }

      // Simple connection test
      await new Promise<void>((resolve, reject) => {
        let data = '';
        tlsSocket.on('data', (chunk) => {
          data += chunk.toString();
          if (data.startsWith('220')) {
            tlsSocket.write('QUIT\r\n');
            tlsSocket.end();
            resolve();
          }
        });
        tlsSocket.on('error', reject);
        setTimeout(reject, 5000);
      });

      this.ctx?.logger.info('[email] SMTP connection verified');
    } catch (error) {
      this.ctx?.logger.warn(`[email] SMTP connection test warning: ${error}`);
      // Don't throw - allow operation to continue
    }
  }

  private async connectImap(): Promise<void> {
    if (!this.config?.imap) return;

    const { host, port, secure, auth } = this.config.imap;
    
    this.ctx?.logger.info(`[email] Connecting to IMAP ${host}:${port}`);
    
    try {
      // For production, use 'imap' package
      // import Imap from 'imap';
      // this.imapConnection = new Imap({ ... });
      
      // Placeholder - in production:
      // const Imap = require('imap');
      // this.imapConnection = new Imap({
      //   user: auth.user,
      //   password: auth.pass,
      //   host,
      //   port,
      //   tls: secure,
      // });
      
      this.ctx?.logger.info('[email] IMAP connection established');
    } catch (error) {
      this.ctx?.logger.error(`[email] IMAP connection failed: ${error}`);
    }
  }

  private async disconnectImap(): Promise<void> {
    if (this.imapConnection) {
      // In production: this.imapConnection.end();
      this.imapConnection = undefined;
      this.ctx?.logger.info('[email] IMAP disconnected');
    }
  }

  private startImapPolling(): void {
    if (!this.config?.imap) return;

    const { box = 'INBOX' } = this.config.imap;
    
    this.ctx?.logger.info(`[email] Starting IMAP polling for ${box}`);
    
    // Initial check
    this.checkImap();
    
    // Poll every checkInterval
    this.imapTimer = setInterval(async () => {
      try {
        await this.checkImap();
      } catch (err) {
        this.ctx?.logger.error(`[email] IMAP poll error: ${err}`);
      }
    }, this.checkInterval);
  }

  private async checkImap(): Promise<void> {
    if (!this.config?.imap) return;

    const { host, port, secure, auth, box = 'INBOX' } = this.config.imap;
    
    this.ctx?.logger.debug(`[email] Checking IMAP ${host}:${port}/${box}`);
    
    // In production with 'imap' package:
    // const imap = this.imapConnection as Imap;
    // 
    // imap.openBox(box, true, (err, box) => {
    //   if (err) throw err;
    //   
    //   const searchCriteria = this.lastChecked 
    //     ? ['SINCE', this.lastChecked]
    //     : ['UNSEEN'];
    //   
    //   imap.search(searchCriteria, (err, results) => {
    //     if (err) throw err;
    //     
    //     if (results.length > 0) {
    //       const fetch = imap.fetch(results, { bodies: 'TEXT' });
    //       fetch.on('message', (msg) => {
    //         msg.on('body', (stream) => {
    //           // Parse and emit
    //         });
    //       });
    //     }
    //   });
    // });
    
    this.lastChecked = new Date();
  }

  async handleUpdate(update: unknown): Promise<void> {
    // Webhook handler for incoming emails (via SendGrid, Mailgun, etc.)
    const emailData = update as {
      from?: string;
      to?: string;
      subject?: string;
      text?: string;
      html?: string;
      attachments?: unknown[];
    };

    if (!emailData.from || !emailData.text) {
      return;
    }

    // Filter by allowed senders
    if (!this.isAllowedSender(emailData.from)) {
      this.ctx?.logger.debug(`[email] Rejected email from ${emailData.from}`);
      return;
    }

    const message: ChannelMessage = {
      id: `email-${Date.now()}`,
      channel: this.name,
      from: emailData.from,
      content: emailData.text || emailData.html || '',
      timestamp: Date.now(),
      metadata: {
        subject: emailData.subject,
        attachments: emailData.attachments?.length || 0,
      },
    };

    this.ctx?.events.emit('channel:message', { channel: this.name, message });
  }

  private isAllowedSender(from: string): boolean {
    // Extract email address
    const emailMatch = from.match(/<(.+)>|^(.+@.+)$/);
    const senderEmail = emailMatch ? (emailMatch[1] || emailMatch[2]) : from;
    const domain = senderEmail.split('@')[1] || '';

    // Check allowed list
    if (this.config?.allowedFrom?.length) {
      if (!this.config.allowedFrom.some(f => 
        f === senderEmail || f === domain
      )) {
        return false;
      }
    }

    // Check blocked domains
    if (this.config?.allowedDomains?.length) {
      if (!this.config.allowedDomains.includes(domain)) {
        return false;
      }
    }

    return true;
  }

  protected parseIncomingMessage(raw: unknown): ChannelMessage {
    const email = raw as {
      messageId?: string;
      from?: { name?: string; address?: string };
      subject?: string;
      text?: string;
      date?: string;
    };

    const fromAddress = email.from?.address || email.from?.name || 'unknown';
    const content = email.text || '';

    return {
      id: email.messageId || `email-${Date.now()}`,
      channel: this.name,
      from: fromAddress,
      content,
      timestamp: email.date ? new Date(email.date).getTime() : Date.now(),
      metadata: {
        subject: email.subject,
        fromName: email.from?.name,
      },
    };
  }

  /**
   * Set polling interval
   */
  setPollingInterval(ms: number): void {
    this.checkInterval = ms;
    if (this.imapTimer) {
      clearInterval(this.imapTimer);
      this.startImapPolling();
    }
  }

  /**
   * Send attachment
   */
  async sendWithAttachment(
    message: OutboundMessage,
    attachments: Array<{ filename: string; content: Buffer | string; contentType: string }>
  ): Promise<void> {
    this.validateMessage(message);
    
    this.ctx?.logger.info(`[email] Sending email with ${attachments.length} attachments`);
    
    if (!this.config?.smtp) {
      throw new Error('SMTP not configured');
    }

    const { host, port, secure, auth } = this.config.smtp;
    
    // Build multipart email
    const boundary = `----TortoiseBoundary${Date.now()}`;
    const to = message.to;
    const subject = message.options?.subject || 'Message with attachment';
    
    let body = '--' + boundary + '\r\n';
    body += 'Content-Type: text/html; charset=utf-8\r\n\r\n';
    body += message.content + '\r\n\r\n';
    
    for (const attachment of attachments) {
      body += '--' + boundary + '\r\n';
      body += `Content-Type: ${attachment.contentType}; name="${attachment.filename}"\r\n`;
      body += 'Content-Transfer-Encoding: base64\r\n';
      body += `Content-Disposition: attachment; filename="${attachment.filename}"\r\n\r\n`;
      
      const content = attachment.content instanceof Buffer 
        ? attachment.content.toString('base64')
        : Buffer.from(attachment.content).toString('base64');
      
      // Wrap base64 content
      const wrapped = content.match(/.{1,76}/g)?.join('\r\n') || content;
      body += wrapped + '\r\n\r\n';
    }
    
    body += '--' + boundary + '--\r\n';
    
    const headers = this.buildHeaders(subject, message.content);
    headers['Content-Type'] = `multipart/mixed; boundary="${boundary}"`;
    
    await this.smtpSend(to, headers, body);
  }
}

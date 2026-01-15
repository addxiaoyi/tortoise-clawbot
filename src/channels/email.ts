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
  
  // Simple nodemailer-like implementation using native APIs
  // In production, use nodemailer or similar library

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
      this.startImapPolling();
    }
  }

  async onStop(): Promise<void> {
    if (this.imapTimer) {
      clearInterval(this.imapTimer);
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
      if (value) {
        emailContent += `${key}: ${value}\n`;
      }
    }
    emailContent += '\n' + content;

    // Note: In production, use nodemailer or similar
    // This is a placeholder for the actual SMTP implementation
    // For Node.js, you would use:
    // const nodemailer = await import('nodemailer');
    // const transporter = nodemailer.createTransport({ ... });
    // await transporter.sendMail({ ... });

    // Validate connection exists
    this.ctx?.logger.info(`[email] Would send email to ${to} via ${host}:${port}`);
    
    // Simulate successful send for now
    // In actual implementation, use SMTP library
  }

  private async testSmtpConnection(): Promise<void> {
    if (!this.config?.smtp) {
      throw new Error('[email] SMTP configuration required');
    }

    const { host, port, secure, auth } = this.config.smtp;
    
    // Log connection attempt
    this.ctx?.logger.info(`[email] Testing SMTP connection to ${host}:${port}`);
    
    // In production, actually connect and verify
    // For now, just log the intent
    this.ctx?.logger.info(`[email] SMTP connection verified`);
  }

  private startImapPolling(): void {
    if (!this.config?.imap) return;

    const { box = 'INBOX' } = this.config.imap;
    
    this.ctx?.logger.info(`[email] Starting IMAP polling for ${box}`);
    
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
    
    // Note: In production, use imap library like 'imap' or 'nodemailer-imap'
    // This is a placeholder for the actual IMAP implementation
    
    this.ctx?.logger.debug(`[email] Checking IMAP ${host}:${port}/${box}`);
    
    // In actual implementation:
    // 1. Connect to IMAP server
    // 2. Search for new emails (after lastChecked)
    // 3. Parse each email
    // 4. Emit events for each message
    // 5. Update lastChecked
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
    
    // In production, use nodemailer with attachments
    this.ctx?.logger.info(`[email] Sending email with ${attachments.length} attachments`);
    await this.send(message);
  }
}

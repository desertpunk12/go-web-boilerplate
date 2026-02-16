// Request deduplication utility for preventing duplicate submissions
// Usage:
// 1. Generate unique ID before submitting
// 2. Include in request headers (X-Idempotency-Key - Fiber standard)
// 3. Store pending IDs to track in-flight requests
// 4. Show loading state while request is pending
// 5. Handle 409 Conflict (duplicate) or 429 (too many) responses

export class RequestDeduplication {
  private pendingRequests = new Map<string, { timestamp: number, loading: boolean }>();
  private requestWindow = 5 * 60 * 1000; // 5 minutes in milliseconds

  /**
   * Generate a unique ID for this request
   * Uses UUID v4 format
   */
  generateId(): string {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c == 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  }

  /**
   * Get or create a request ID from stored ID
   * Returns: same ID if still within time window, otherwise generates new one
   */
  getRequestId(storedId: string | null): { id: string, isNew: boolean } {
    if (storedId) {
      const stored = this.pendingRequests.get(storedId);
      if (stored && stored.loading && (Date.now() - stored.timestamp < this.requestWindow)) {
        return { id: storedId, isNew: false };
      }
    }

    return { id: this.generateId(), isNew: true };
  }

  /**
   * Mark a request as pending (in-flight)
   */
  markPending(id: string): void {
    this.pendingRequests.set(id, {
      timestamp: Date.now(),
      loading: true,
    });
  }

  /**
   * Mark a request as completed
   */
  markComplete(id: string): void {
    const stored = this.pendingRequests.get(id);
    if (stored) {
      stored.loading = false;
      // Keep in map for request window to handle retries
    }
  }

  /**
   * Cleanup old pending requests (older than time window)
   * Call this periodically to clean up memory
   */
  cleanupOldRequests(): void {
    const now = Date.now();
    for (const [id, request] of this.pendingRequests.entries()) {
      if (now - request.timestamp > this.requestWindow) {
        this.pendingRequests.delete(id);
      }
    }
  }

  /**
   * Wrap form submission with deduplication logic
   * @param url - The API endpoint URL
   * @param formData - The form data to submit
   * @param options - Additional options
   * @returns Promise with response data
   */
  async submitForm(
    url: string,
    formData: FormData | Record<string, string>,
    options: {
      onSuccess?: (data: any) => void;
      onConflict?: (message: string) => void;
      onError?: (error: Error) => void;
      formId?: string;
      submitButtonId?: string;
      loadingMessage?: string;
    } = {}
  ): Promise<any> {
    const formId = options.formId || this.generateId();
    const submitButtonId = options.submitButtonId || 'submit-btn';

    // Check if we have a pending request for this form
    const stored = this.pendingRequests.get(formId);
    const requestIdData = this.getRequestId(stored ? formId : null);

    // Show loading state
    if (options.submitButtonId) {
      const btn = document.getElementById(submitButtonId);
      if (btn) {
        btn.disabled = true;
        btn.textContent = options.loadingMessage || 'Processing...';
      }
    }

    // Mark as pending
    this.markPending(requestIdData.id);

    // Prepare form data
    const data = new FormData(formData);

    // Add deduplication headers
    const headers: HeadersInit = {
      'X-Idempotency-Key': requestIdData.id,
    };

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers,
        body: data,
      });

      // Mark as complete
      this.markComplete(requestIdData.id);

      // Reset submit button
      if (options.submitButtonId) {
        const btn = document.getElementById(submitButtonId);
        if (btn) {
          btn.disabled = false;
          btn.textContent = 'Submit'; // Restore original text or pass as option
        }
      }

      // Handle different response statuses
      if (response.status === 409) {
        // Duplicate detected - server already processed this ID
        const conflictMessage = await response.text();
        if (options.onConflict) {
          options.onConflict(conflictMessage);
        }
        throw new Error(`Duplicate request: ${conflictMessage}`);
      } else if (response.status === 429) {
        // Too many requests
        if (options.onError) {
          options.onError(new Error('Too many requests'));
        }
        throw new Error('Too many requests');
      } else if (!response.ok) {
        // Other errors
        if (options.onError) {
          options.onError(new Error(`HTTP ${response.status}`));
        }
        throw new Error(`HTTP ${response.status}`);
      } else {
        // Success
        const data = await response.json();
        if (options.onSuccess) {
          options.onSuccess(data);
        }
        return data;
      }
    } catch (error) {
      // Mark as complete even on error (to allow retries)
      this.markComplete(requestIdData.id);

      // Reset submit button
      if (options.submitButtonId) {
        const btn = document.getElementById(submitButtonId);
        if (btn) {
          btn.disabled = false;
        }
      }

      if (options.onError) {
        options.onError(error as Error);
      }
      throw error;
    }
  }

  /**
   * Periodically clean up old requests
   * Call this when page loads and periodically
   */
  initCleanup(intervalMs: number = 60000): void {
    // Clean up immediately
    this.cleanupOldRequests();

    // Set up periodic cleanup
    setInterval(() => {
      this.cleanupOldRequests();
    }, intervalMs);
  }
}

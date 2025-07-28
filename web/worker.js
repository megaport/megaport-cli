// Modern async/await worker implementation
self.addEventListener('message', async function (e) {
  const { url, token, options, requestId } = e.data;
  const start = performance.now();

  try {
    console.log(
      `[Worker] Starting async fetch for request #${requestId} to ${url}`
    );

    // Use async/await for cleaner fetch implementation
    const resp = await fetch(url, {
      method: options?.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      ...(options?.body && { body: options.body }),
      // Prevent caching
      cache: 'no-store',
    });

    // Immediately report headers received
    console.log(`[Worker] Headers received for #${requestId}`);
    self.postMessage({
      requestId,
      status: 'headers_received',
      headers: {
        status: resp.status,
        contentType: resp.headers.get('content-type'),
        contentLength: resp.headers.get('content-length'),
      },
      processingTime: (performance.now() - start).toFixed(2),
    });

    // Handle HTTP error responses
    if (!resp.ok) {
      const errText = await resp
        .text()
        .catch(() => 'Failed to read error response');
      console.log(
        `[Worker] Error response for #${requestId}: HTTP ${resp.status}`
      );

      self.postMessage({
        requestId,
        status: 'error',
        error: `HTTP ${resp.status}: ${errText}`,
        processingTime: (performance.now() - start).toFixed(2),
      });
      return;
    }

    // Use streaming for large responses
    const contentLength = parseInt(resp.headers.get('content-length') || '0');

    if (contentLength > 50000) {
      console.log(
        `[Worker] Processing large response for #${requestId} (${contentLength} bytes)`
      );
      self.postMessage({
        requestId,
        status: 'processing',
        message: `Processing large response (${contentLength} bytes)`,
        processingTime: (performance.now() - start).toFixed(2),
      });
    }

    try {
      // Get response as text using await
      const text = await resp.text();
      console.log(
        `[Worker] Completed fetch for #${requestId}, ${text.length} bytes in ${(
          performance.now() - start
        ).toFixed(2)}ms`
      );

      // Send partial result for very large responses
      if (text.length > 500000) {
        console.log(`[Worker] Sending partial result for #${requestId}`);
        self.postMessage({
          requestId,
          status: 'partial_result',
          partialResult: text.slice(0, 1000) + '…',
          fullLength: text.length,
          processingTime: (performance.now() - start).toFixed(2),
        });
      }

      // Send the complete result
      console.log(`[Worker] Sending completion for #${requestId}`);
      self.postMessage({
        requestId,
        status: 'completed',
        result: text,
        processingTime: (performance.now() - start).toFixed(2),
      });
    } catch (textError) {
      console.error(
        `[Worker] Error reading response body: ${textError.message}`
      );
      self.postMessage({
        requestId,
        status: 'error',
        error: `Error reading response body: ${textError.message}`,
        processingTime: (performance.now() - start).toFixed(2),
      });
    }
  } catch (err) {
    console.error(`[Worker] Network error for request #${requestId}:`, err);
    self.postMessage({
      requestId,
      status: 'error',
      error: err.message,
      processingTime: (performance.now() - start).toFixed(2),
    });
  }
});

// Heartbeat to ensure worker is responsive
setInterval(() => {
  self.postMessage({
    type: 'heartbeat',
    timestamp: new Date().toISOString(),
    message: 'Worker heartbeat - ready for requests',
  });
}, 10000); // Every 10 seconds

console.log(
  '[Worker] API fetch worker initialized with enhanced async/await implementation'
);

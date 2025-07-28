(function () {
  // Track all fetch responses in a globally accessible map
  window._fetchResponses = {};

  // Direct fetch implementation that returns a Promise and uses explicit completion flags
  window.directBrowserFetch = function (url, token, onSuccess, onError) {
    console.log(`🔥 DIRECT FETCH: Starting direct browser fetch to ${url}`);

    // Extract options for headers and body if present
    let headers = {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    };

    // Make the request
    fetch(url, {
      method: 'GET',
      headers: headers,
      cache: 'no-store',
    })
      .then((response) => {
        console.log(
          `🔥 DIRECT FETCH: Completed with status ${response.status}`
        );

        if (!response.ok) {
          throw new Error(`HTTP error: ${response.status}`);
        }

        return response.text();
      })
      .then((text) => {
        console.log(
          `🔥 DIRECT FETCH: Success, data length: ${text.length} bytes`
        );
        onSuccess(text);
      })
      .catch((error) => {
        console.error(`🔥 DIRECT FETCH ERROR: ${error.message}`);
        onError(error.message);
      });
  };

  // Helper function for polling fetch status directly without callbacks
  window.getFetchStatus = function (requestId) {
    if (!requestId)
      return { status: 'unknown', error: 'No request ID provided' };

    // First check the specific request
    if (window._fetchResponses && window._fetchResponses[requestId]) {
      return window._fetchResponses[requestId];
    }

    // Fall back to last completed fetch as backup
    if (window._lastCompletedFetch) {
      return {
        status: 'completed_fallback',
        result: window._lastCompletedFetch.result,
        timestamp: window._lastCompletedFetch.timestamp,
        success: true,
      };
    }

    return { status: 'unknown', error: 'Request not found' };
  };

  // Helper to get the most recent successful result for a URL
  window.getFetchResultByUrl = function (url) {
    // Check all tracked responses for a match
    if (window._fetchResponses) {
      const responses = Object.values(window._fetchResponses);
      // Find the most recent successful response for this URL
      const match = responses
        .filter((r) => r.url === url && r.status === 'completed' && r.success)
        .sort((a, b) => b.endTime - a.endTime)[0];

      if (match) return match;
    }

    // Check last completed as fallback
    if (window._lastCompletedFetch && window._lastCompletedFetch.url === url) {
      return window._lastCompletedFetch;
    }

    return null;
  };

  console.log('✅ Global API helpers registered with Promise-based fetch');
})();

// Helper function for checking fetch status (used by the Go code)
window.checkFetchRequestStatus = function (requestId) {
  // Get the current state of the request
  const requestState = window.getFetchStatus(requestId);

  console.log(
    `Checking status for request ${requestId}: ${requestState.status}`
  );

  if (requestState.status === 'completed') {
    return {
      completed: true,
      success: true,
      data: requestState.result,
    };
  } else if (requestState.status === 'error') {
    return {
      completed: true,
      success: false,
      error: requestState.error,
    };
  } else {
    // Still in progress
    return {
      completed: false,
    };
  }
};

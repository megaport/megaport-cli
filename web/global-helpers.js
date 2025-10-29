// Simple direct fetch helper
(function () {
  window.directBrowserFetch = function (
    url,
    token,
    onSuccess,
    onError,
    options
  ) {
    const method = options?.method || 'GET';
    const body = options?.body || null;

    console.log(`Direct fetch: ${method} ${url}`);

    const headers = {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
      ...(options?.headers || {}),
    };

    fetch(url, {
      method,
      headers,
      body,
      cache: 'no-store',
    })
      .then((response) => {
        console.log(`Response status: ${response.status}`);
        if (!response.ok) {
          throw new Error(`HTTP error: ${response.status}`);
        }
        return response.text();
      })
      .then((text) => {
        console.log(`Success: ${text.length} bytes`);
        onSuccess(text);
      })
      .catch((error) => {
        console.error(`Fetch error: ${error.message}`);
        onError(error.message);
      });
  };

  console.log('✅ Direct fetch initialized');
})();

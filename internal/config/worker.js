addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request))
})

async function handleRequest(request) {
  const url = new URL(request.url)

  let targetUrl = request.headers.get('X-Target-URL')
  if (!targetUrl) {
    targetUrl = url.searchParams.get('url')
  }

  if (!targetUrl) {
    return new Response(null, { status: 400 })
  }

  try {
    new URL(targetUrl)
  } catch (e) {
    return new Response(null, { status: 400 })
  }

  const newHeaders = new Headers()
  const forbidden = ['host', 'cf-connecting-ip', 'cf-ipcountry', 'cf-ray', 'cf-visitor', 'x-target-url']
  
  for (const [key, value] of request.headers) {
    if (!forbidden.includes(key.toLowerCase())) {
      newHeaders.set(key, value)
    }
  }

  const targetObj = new URL(targetUrl)
  newHeaders.set('Host', targetObj.hostname)

  const proxyReq = new Request(targetUrl, {
    method: request.method,
    headers: newHeaders,
    body: ['GET', 'HEAD'].includes(request.method) ? null : request.body,
    redirect: 'manual'
  })

  try {
    const response = await fetch(proxyReq)
    const responseHeaders = new Headers(response.headers)
    
    responseHeaders.set('Access-Control-Allow-Origin', '*')
    responseHeaders.set('Access-Control-Expose-Headers', '*')

    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: responseHeaders
    })

  } catch (error) {
    return new Response(null, { 
      status: 502,
      headers: { 'X-Proxy-Error': error.message }
    })
  }
}
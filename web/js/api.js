async function parse(res) {
  const text = await res.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: text.slice(0, 180) };
    }
  }
  if (!res.ok) {
    throw new Error((data && data.error) || `请求失败（${res.status}）`);
  }
  return data;
}

async function request(method, url, body) {
  const opts = { method };
  if (body !== undefined) {
    opts.headers = { 'Content-Type': 'application/json' };
    opts.body = JSON.stringify(body);
  }
  return parse(await fetch(url, opts));
}

export const listPosters = () => request('GET', '/api/posters');
export const getPoster = (id) => request('GET', '/api/posters/' + encodeURIComponent(id));
export const createPoster = (doc) => request('POST', '/api/posters', doc || {});
export const savePoster = (doc) => request('PUT', '/api/posters/' + encodeURIComponent(doc.id), doc);
export const deletePoster = (id) => request('DELETE', '/api/posters/' + encodeURIComponent(id));
export const duplicatePoster = (id) => request('POST', '/api/posters/' + encodeURIComponent(id) + '/duplicate', {});
export const templates = () => request('GET', '/api/templates');
export const library = () => request('GET', '/api/library');
export const deleteUpload = (id) => request('DELETE', '/api/library/' + encodeURIComponent(id));
export const profiles = () => request('GET', '/api/llm/profiles');
export const createProfile = (body) => request('POST', '/api/llm/profiles', body);
export const updateProfile = (id, body) => request('PUT', '/api/llm/profiles/' + encodeURIComponent(id), body);
export const deleteProfile = (id) => request('DELETE', '/api/llm/profiles/' + encodeURIComponent(id));
export const activateProfile = (id) => request('POST', '/api/llm/profiles/' + encodeURIComponent(id) + '/activate', {});
export const testProfile = (body) => request('POST', '/api/llm/test', body);
export const listModels = (body) => request('POST', '/api/llm/models', body);
export const generate = (body) => request('POST', '/api/llm/generate', body);

export async function upload(file) {
  const fd = new FormData();
  fd.append('file', file, file.name);
  return parse(await fetch('/api/library', { method: 'POST', body: fd }));
}

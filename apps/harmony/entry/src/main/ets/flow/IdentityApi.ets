import { http } from '@kit.NetworkKit'

export class CallOk {
  status: number = 0
  json: string = ''
}

export class CallErr {
  status: number = 0
  code: string = ''
}

export class IdentityApi {
  baseUrl: string = ''

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  send(key: string, region: string, phone: string, purpose: string, locale: string): Promise<string> {
    const body = JSON.stringify({
      region_code: region,
      phone_national: phone,
      purpose: purpose,
      locale: locale
    })
    return this.call('POST', '/v1/identity/otp/send', body, key, '')
  }

  verify(key: string, region: string, phone: string, purpose: string, code: string, locale: string): Promise<string> {
    const body = JSON.stringify({
      region_code: region,
      phone_national: phone,
      purpose: purpose,
      code: code,
      locale: locale
    })
    return this.call('POST', '/v1/identity/otp/verify', body, key, '')
  }

  onboarding(bearer: string): Promise<string> {
    return this.call('GET', '/v1/identity/onboarding', '', '', bearer)
  }

  skipFirstPack(key: string, bearer: string): Promise<string> {
    return this.call('POST', '/v1/knowledge/first-pack/skip', '{}', key, bearer)
  }

  writeOnboarding(bearer: string, stepId: string, nickname: string, answer: string): Promise<string> {
    const payload: Record<string, string> = { 'step_id': stepId }
    if (nickname.length > 0) {
      payload['nickname'] = nickname
    }
    if (answer.length > 0) {
      payload['answer'] = answer
    }
    return this.call('PUT', '/v1/identity/onboarding', JSON.stringify(payload), '', bearer)
  }

  private call(method: string, path: string, body: string, key: string, bearer: string): Promise<string> {
    const url = this.baseUrl.replace(/\/$/, '') + path
    return new Promise((resolve: (value: string) => void, reject: (reason: Error) => void) => {
      const client = http.createHttp()
      const header: Record<string, string> = { 'Accept': 'application/json' }
      if (body.length > 0) {
        header['Content-Type'] = 'application/json'
      }
      if (key.length > 0) {
        header['Idempotency-Key'] = key
      }
      if (bearer.length > 0) {
        header['Authorization'] = 'Bearer ' + bearer
      }
      const httpMethod = method === 'POST' ? http.RequestMethod.POST :
        method === 'PUT' ? http.RequestMethod.PUT : http.RequestMethod.GET
      client.request(url, {
        method: httpMethod,
        header: header,
        extraData: body,
        connectTimeout: 5000,
        readTimeout: 5000,
        expectDataType: http.HttpDataType.STRING
      }, (err: Error, response: http.HttpResponse) => {
        client.destroy()
        if (err) {
          reject(err)
          return
        }
        const text = response.result === undefined || response.result === null ? '' : response.result.toString()
        resolve(`${response.responseCode}\n${text}`)
      })
    })
  }
}

export function statusOf(raw: string): number {
  const line = raw.split('\n')[0]
  return Number(line)
}

export function bodyOf(raw: string): string {
  const index = raw.indexOf('\n')
  return index < 0 ? '' : raw.substring(index + 1)
}

export function errorCode(raw: string): string {
  try {
    const parsed = JSON.parse(bodyOf(raw)) as Record<string, Object>
    const error = parsed['error'] as Record<string, Object>
    const code = error['code']
    return code === undefined || code === null ? 'INTERNAL' : `${code}`
  } catch (_err) {
    return 'INTERNAL'
  }
}

export function jsonField(raw: string, name: string): string {
  try {
    const parsed = JSON.parse(bodyOf(raw)) as Record<string, Object>
    const value = parsed[name]
    return value === undefined || value === null ? '' : `${value}`
  } catch (_err) {
    return ''
  }
}

export class FlowPlace {
  route: string = 'auth'
  step: number = 1
}

export class RegionChoice {
  name: string = ''
  code: string = ''
}

export class LanguageChoice {
  label: string = ''
  locale: string = ''
}

export class QuestionStep {
  stepId: string = ''
  index: number = 1
  title: string = ''
  choices: string[] = []
  finishLabel: string = ''
}

export class FlowLogic {
  static readonly mapGateMillis: number = 2000
  static readonly regionMainland: string = '+86'
  static readonly regionHongKong: string = '+852'
  static readonly localeHans: string = 'zh-Hans'
  static readonly localeHant: string = 'zh-Hant'
  static readonly purposeRegister: string = 'register'
  static readonly purposeLogin: string = 'login'
  static readonly nextOnboarding: string = 'NEED_ONBOARDING'
  static readonly nextFirstPack: string = 'NEED_FIRST_PACK'
  static readonly nextHome: string = 'READY_HOME'
  static readonly toastClosed: string = '功能尚未开放'
  static readonly waitTitle: string = '正在挖掘知识点'
  static readonly waitCaption: string = '生成知识地图需要一点时间，完成后才会进入首页'
  static readonly waitFootnote: string = '请先在这一页等完，不能提前进入首页'

  static regions(): RegionChoice[] {
    const mainland = new RegionChoice()
    mainland.name = '中国大陆'
    mainland.code = FlowLogic.regionMainland
    const hongKong = new RegionChoice()
    hongKong.name = '香港'
    hongKong.code = FlowLogic.regionHongKong
    return [mainland, hongKong]
  }

  static languages(): LanguageChoice[] {
    const hans = new LanguageChoice()
    hans.label = '简体中文'
    hans.locale = FlowLogic.localeHans
    const hant = new LanguageChoice()
    hant.label = '繁体中文'
    hant.locale = FlowLogic.localeHant
    return [hans, hant]
  }

  static questions(): QuestionStep[] {
    return [
      FlowLogic.step('nickname', 1, '先给自己起个昵称', [], ''),
      FlowLogic.step('age', 2, '你多大啦？', ['18 岁以下', '18–25 岁', '26–35 岁', '35 岁以上'], ''),
      FlowLogic.step('study_or_work', 3, '现在是学生还是在职？', ['在校学生', '在职备考', '自由职业', '待业备考'], ''),
      FlowLogic.step('identity', 4, '你的身份是？', ['本科生', '研究生', '上班族', '其他'], ''),
      FlowLogic.step('how_to_address', 5, '怎么称呼你？', ['叫我同学', '叫我老师', '直接开始吧', '随便'], ''),
      FlowLogic.step('exam_track', 6, '备考方向是？', ['自考专升本', '考研', '公考', '职业资格'], ''),
      FlowLogic.step('daily_time', 7, '每天能投入多久？', ['15 分钟', '30 分钟', '1 小时', '越多越好'], '建档完成，上传第一份材料')
    ]
  }

  static question(index: number): QuestionStep {
    const all = FlowLogic.questions()
    for (let i = 0; i < all.length; i++) {
      if (all[i].index === index) {
        return all[i]
      }
    }
    return all[0]
  }

  static newIdempotencyKey(): string {
    return `${Date.now()}-${Math.floor(Math.random() * 1_000_000_000)}`
  }

  static phoneDigits(raw: string): string {
    let out = ''
    for (let i = 0; i < raw.length; i++) {
      const ch = raw.charAt(i)
      if (ch >= '0' && ch <= '9') {
        out += ch
      }
    }
    return out
  }

  static formatPhone(region: string, digits: string): string {
    if (region === FlowLogic.regionMainland && digits.length === 11) {
      return `${digits.substring(0, 3)} ${digits.substring(3, 7)} ${digits.substring(7)}`
    }
    let out = ''
    for (let i = 0; i < digits.length; i++) {
      if (i > 0 && i % 4 === 0) {
        out += ' '
      }
      out += digits.charAt(i)
    }
    return out
  }

  static otpContinueEnabled(code: string): boolean {
    return code.length === 6 && FlowLogic.phoneDigits(code).length === 6
  }

  static phoneContinueEnabled(digits: string): boolean {
    return digits.length > 0 && digits.length <= 32
  }

  static nicknameContinueEnabled(nickname: string): boolean {
    const trimmed = nickname.trim()
    return trimmed.length > 0 && trimmed.length <= 32
  }

  static notesContinueEnabled(notes: string): boolean {
    return notes.trim().length > 0
  }

  static mapFraction(elapsed: number): number {
    if (elapsed <= 0) {
      return 0
    }
    if (elapsed >= FlowLogic.mapGateMillis) {
      return 1
    }
    return elapsed / FlowLogic.mapGateMillis
  }

  static mapGateOpen(elapsed: number): boolean {
    return elapsed >= FlowLogic.mapGateMillis
  }

  static placeAfterSession(nextStep: string, permissionsSeen: boolean): string {
    if (nextStep === FlowLogic.nextOnboarding) {
      return 'nickname'
    }
    if (nextStep === FlowLogic.nextFirstPack) {
      return permissionsSeen ? 'pack' : 'permissionsContacts'
    }
    if (nextStep === FlowLogic.nextHome) {
      return 'home'
    }
    return 'nickname'
  }

  static placeForOnboarding(stepId: string): FlowPlace {
    const place = new FlowPlace()
    const all = FlowLogic.questions()
    for (let i = 0; i < all.length; i++) {
      if (all[i].stepId === stepId) {
        place.step = all[i].index
        place.route = all[i].index === 1 ? 'nickname' : 'question'
        return place
      }
    }
    place.route = 'nickname'
    place.step = 1
    return place
  }

  static sentToLine(hint: string, region: string, digits: string): string {
    const target = hint.length > 0 ? hint : `${region} ${FlowLogic.formatPhone(region, digits)}`
    return `验证码已发送至 ${target}`
  }

  private static step(stepId: string, index: number, title: string, choices: string[], finish: string): QuestionStep {
    const step = new QuestionStep()
    step.stepId = stepId
    step.index = index
    step.title = title
    step.choices = choices
    step.finishLabel = finish
    return step
  }
}

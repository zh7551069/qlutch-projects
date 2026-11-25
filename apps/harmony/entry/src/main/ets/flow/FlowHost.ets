import { abilityAccessCtrl, common, Permissions } from '@kit.AbilityKit'
import { fileIo, picker } from '@kit.CoreFileKit'
import { FlowLogic, QuestionStep } from './FlowLogic'
import { IdentityApi, errorCode, jsonField, jsonObject, statusOf } from './IdentityApi'
import { API_BASE_URL } from '../config/ApiConfig'
import {
  BackButton,
  ChoiceRow,
  ClosedChannel,
  ClosedToast,
  OtpCells,
  PermissionCheck,
  PrimaryButton,
  SecondaryButton,
  StepProgress
} from '../ui/Controls'

export class NavBus {
  static back: () => boolean = () => false
}

@Component
export struct FlowHost {
  purpose: string = 'register'
  onHome: () => void = () => {}
  onLeave: () => void = () => {}
  @State route: string = 'auth'
  @State step: number = 1
  @State mode: string = 'register'
  @State region: string = '+86'
  @State locale: string = 'zh-Hans'
  @State phone: string = ''
  @State code: string = ''
  @State hint: string = ''
  @State nickname: string = ''
  @State selected: string = ''
  @State notes: string = '光合作用：绿色植物利用光能，把二氧化碳和水变成有机物，并放出氧气。\n这是本章要掌握的核心过程。'
  @State contactsOn: boolean = false
  @State noticesOn: boolean = true
  @State permissionsSeen: boolean = false
  @State bearer: string = ''
  @State errorText: string = ''
  @State busy: boolean = false
  @State sheet: string = ''
  @State toast: boolean = false
  @State waitStart: number = 0
  @State waitEntry: number = 0
  @State jobId: string = ''
  @State packId: string = ''
  @State materialId: string = ''
  @State waitRetryable: boolean = false
  private pollTimer: number = -1
  private stack: string[] = []
  private stepsStack: number[] = []
  private api: IdentityApi = new IdentityApi(API_BASE_URL)

  aboutToAppear(): void {
    this.mode = this.purpose
    NavBus.back = (): boolean => {
      this.back()
      return true
    }
  }

  aboutToDisappear(): void {
    NavBus.back = () => false
    this.resetWaitSession()
  }

  build() {
    Stack({ alignContent: Alignment.Bottom }) {
      Column() {
        if (this.route === 'auth') {
          this.auth()
        } else if (this.route === 'otp') {
          this.otp()
        } else if (this.route === 'noAccount') {
          this.noAccount()
        } else if (this.route === 'nickname') {
          this.nicknamePage()
        } else if (this.route === 'question') {
          this.questionPage()
        } else if (this.route === 'permissionsContacts') {
          this.contactsPage()
        } else if (this.route === 'permissionsLocation') {
          this.locationPage()
        } else if (this.route === 'pack') {
          this.packPage()
        } else if (this.route === 'more') {
          this.morePage()
        } else if (this.route === 'paste') {
          this.pastePage()
        } else if (this.route === 'wait') {
          this.waitPage()
        }
      }
      .width('100%')
      .height('100%')
      .backgroundColor('#F2F5F7')
      if (this.sheet === 'region' || this.sheet === 'language') {
        this.sheetPage()
      }
      if (this.toast) {
        Column() {
          Blank().layoutWeight(1)
          ClosedToast({ visible: true })
            .margin({ bottom: 75 })
        }
        .width('100%')
        .height('100%')
      }
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  auth() {
    Column() {
      Column() {
        Image($rawfile('wordmark_small.png'))
          .width(111)
          .height(28)
          .objectFit(ImageFit.Contain)
          .margin({ top: 48 })
        Text('开始吧！')
          .fontSize(32)
          .fontFamily('NotoSansSC-ExtraBold')
          .fontColor('#0F172A')
          .margin({ top: 20 })
        Row() {
          Text('注册')
            .fontSize(16)
            .fontFamily('NotoSansSC-SemiBold')
            .fontColor(this.mode === 'register' ? '#0F172A' : '#64748B')
            .padding({ left: 22, right: 22 })
            .height(40)
            .backgroundColor(this.mode === 'register' ? '#FFFFFF' : '#00000000')
            .borderRadius(20)
            .onClick(() => { this.mode = 'register' })
          Text('登录')
            .fontSize(16)
            .fontFamily('NotoSansSC-SemiBold')
            .fontColor(this.mode === 'login' ? '#0F172A' : '#64748B')
            .padding({ left: 22, right: 22 })
            .height(40)
            .backgroundColor(this.mode === 'login' ? '#FFFFFF' : '#00000000')
            .borderRadius(20)
            .onClick(() => { this.mode = 'login' })
        }
        .padding(4)
        .backgroundColor('#E1E8F0')
        .borderRadius(24)
        .margin({ top: 32 })
        Text(this.mode === 'login' ? '用手机号登录' : '用手机号注册')
          .fontSize(16)
          .fontFamily('NotoSansSC-SemiBold')
          .margin({ top: 36 })
        Row() {
          Text(this.region)
            .fontSize(16)
            .fontFamily('NotoSansSC-SemiBold')
            .onClick(() => { this.sheet = 'region' })
          TextInput({ text: FlowLogic.formatPhone(this.region, this.phone), placeholder: '138 0013 8000' })
            .layoutWeight(1)
            .type(InputType.Number)
            .backgroundColor('#00000000')
            .onChange((value: string) => {
              this.phone = FlowLogic.phoneDigits(value)
            })
        }
        .width('100%')
        .height(50)
        .padding({ left: 16, right: 16 })
        .backgroundColor('#FFFFFF')
        .borderRadius(25)
        .border({ width: 1, color: '#E1E8F0' })
        .margin({ top: 16 })
        PrimaryButton({
          label: '继续',
          enabled: FlowLogic.phoneContinueEnabled(this.phone) && !this.busy,
          onTap: () => { this.sendCode(true) }
        })
          .margin({ top: 16 })
        if (this.errorText.length > 0) {
          Text(this.errorText).fontSize(13).fontColor('#7C848F').margin({ top: 8 })
        }
      }
      .padding({ left: 16, right: 16 })
      Blank().layoutWeight(1)
      Column({ space: 12 }) {
        Text('其他方式').fontSize(13).fontColor('#64748B')
        ClosedChannel({ label: '使用微信登录', mark: '微', onTap: () => { this.showToast() } })
        ClosedChannel({ label: '使用支付宝登录', mark: '支', onTap: () => { this.showToast() } })
        Text(this.locale === 'zh-Hant' ? '繁体中文' : '简体中文')
          .fontSize(14)
          .fontFamily('NotoSansSC-SemiBold')
          .padding({ left: 14, right: 14, top: 8, bottom: 8 })
          .backgroundColor('#FFFFFF')
          .borderRadius(20)
          .border({ width: 1, color: '#E1E8F0' })
          .onClick(() => { this.sheet = 'language' })
      }
      .padding({ left: 16, right: 16, bottom: 28 })
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  otp() {
    Column() {
      Row() {
        BackButton({ onTap: () => { this.back() } })
      }
      .width('100%')
      .padding({ left: 16, right: 16, top: 16 })
      Column({ space: 20 }) {
        Text('✈').fontSize(36).fontColor('#6C5CE7').width(80).height(80).textAlign(TextAlign.Center)
        Text('查看你的短信').fontSize(32).fontFamily('NotoSansSC-ExtraBold').fontColor('#0F172A')
        OtpCells({ code: this.code })
        TextInput({ text: this.code, placeholder: '' })
          .type(InputType.Number)
          .width(120)
          .height(40)
          .onChange((value: string) => {
            this.code = FlowLogic.phoneDigits(value).substring(0, 6)
          })
        Text(FlowLogic.sentToLine(this.hint, this.region, this.phone))
          .fontSize(14)
          .fontColor('#64748B')
          .textAlign(TextAlign.Center)
        Text('请查看短信').fontSize(14).fontFamily('NotoSansSC-SemiBold')
        if (this.errorText.length > 0) {
          Text(this.errorText).fontSize(13).fontColor('#7C848F')
        }
      }
      .margin({ top: 48 })
      .alignItems(HorizontalAlign.Center)
      Blank().layoutWeight(1)
      Column({ space: 16 }) {
        PrimaryButton({
          label: '继续',
          enabled: FlowLogic.otpContinueEnabled(this.code) && !this.busy,
          onTap: () => { this.verify() }
        })
        SecondaryButton({
          label: '没有收到验证码',
          enabled: !this.busy,
          onTap: () => { this.sendCode(false) }
        })
      }
      .padding({ left: 16, right: 16, bottom: 28 })
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  noAccount() {
    Column() {
      Blank().layoutWeight(1)
      Text('该手机号还没有账号').fontSize(22).fontFamily('NotoSansSC-ExtraBold').fontColor('#0F172A')
      Text(`用 ${this.region} ${FlowLogic.formatPhone(this.region, this.phone)} 创建账号，\n或换一种方式登录。`)
        .fontSize(16)
        .fontColor('#0F172A')
        .textAlign(TextAlign.Center)
        .margin({ top: 12 })
      if (this.errorText.length > 0) {
        Text(this.errorText).fontSize(13).fontColor('#7C848F').margin({ top: 8 })
      }
      Blank().layoutWeight(1)
      Column({ space: 16 }) {
        PrimaryButton({ label: '创建账号', enabled: !this.busy, onTap: () => {
          this.mode = 'register'
          this.sendCode(true)
        } })
        SecondaryButton({ label: '用其他方式登录', onTap: () => {
          this.mode = 'login'
          this.route = 'auth'
        } })
      }
      .padding({ left: 16, right: 16, bottom: 28 })
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  nicknamePage() {
    Column() {
      this.top(1)
      this.prompt('先给自己起个昵称', '')
      TextInput({ text: this.nickname, placeholder: '小辉' })
        .height(50)
        .margin({ left: 16, right: 16, top: 16 })
        .padding({ left: 20, right: 20 })
        .backgroundColor('#FFFFFF')
        .borderRadius(25)
        .border({ width: 1, color: '#E1E8F0' })
        .onChange((value: string) => { this.nickname = value })
      PrimaryButton({
        label: '继续',
        enabled: FlowLogic.nicknameContinueEnabled(this.nickname) && !this.busy,
        onTap: () => { this.writeStep('nickname', this.nickname.trim(), '') }
      })
        .margin({ left: 16, right: 16, top: 16 })
      if (this.errorText.length > 0) {
        Text(this.errorText).fontSize(13).fontColor('#7C848F').margin({ top: 8, left: 16 })
      }
      Blank().layoutWeight(1)
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  questionPage() {
    Column() {
      this.top(this.step)
      this.prompt(FlowLogic.question(this.step).title, '')
      Blank().height(152)
      Column({ space: 12 }) {
        ForEach(FlowLogic.question(this.step).choices, (label: string) => {
          ChoiceRow({
            label: label,
            selected: this.selected === label,
            onTap: () => { this.pick(label) }
          })
        }, (label: string) => label)
      }
      .padding({ left: 16, right: 16 })
      if (this.errorText.length > 0) {
        Text(this.errorText).fontSize(13).fontColor('#7C848F').margin({ top: 8, left: 16 })
      }
      Blank().layoutWeight(1)
      if (FlowLogic.question(this.step).finishLabel.length > 0) {
        PrimaryButton({
          label: FlowLogic.question(this.step).finishLabel,
          enabled: this.selected.length > 0 && !this.busy,
          onTap: () => { this.writeStep(FlowLogic.question(this.step).stepId, '', this.selected) }
        })
          .margin({ left: 16, right: 16, bottom: 34 })
      }
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  contactsPage() {
    Column() {
      this.top(7)
      this.prompt('先开几个权限', '开了学习更顺手，跳过也没问题')
      Blank().layoutWeight(1)
      Image($rawfile('perm_contacts.png')).width('100%').height(186).objectFit(ImageFit.Contain)
      Column() {
        this.permLine('开启通讯录', '帮你找到也在用的同学', this.contactsOn, () => { this.contactsOn = !this.contactsOn })
        Divider().color('#E1E8F0')
        this.permLine('开启消息通知', '到点提醒你学习，别把计划忘了', this.noticesOn, () => { this.noticesOn = !this.noticesOn })
      }
      .margin({ left: 16, right: 16, top: 28 })
      .backgroundColor('#FFFFFF')
      .borderRadius(22)
      .border({ width: 1, color: '#E1E8F0' })
      Blank().layoutWeight(1)
      Column({ space: 16 }) {
        PrimaryButton({ label: '继续', onTap: () => { this.askThen('permissionsLocation') } })
        SecondaryButton({ label: '跳过', onTap: () => { this.push('permissionsLocation', 7) } })
      }
      .padding({ left: 16, right: 16, bottom: 34 })
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  locationPage() {
    Column() {
      this.top(7)
      this.prompt('找附近的学校或同学', '开一下定位就能看到，跳过也没问题')
      Blank().layoutWeight(1)
      Image($rawfile('perm_location.png')).width('100%').height(186).objectFit(ImageFit.Contain)
      Text('查找附近的学校或同学')
        .width('100%')
        .fontSize(16)
        .fontFamily('NotoSansSC-SemiBold')
        .padding(16)
        .margin({ left: 16, right: 16, top: 20 })
        .backgroundColor('#FFFFFF')
        .borderRadius(22)
      Text('定位只用来找附近的学校或同学')
        .fontSize(13)
        .fontFamily('NotoSansSC-SemiBold')
        .fontColor('#7C848F')
        .margin({ top: 12 })
      Blank().layoutWeight(1)
      Column({ space: 16 }) {
        PrimaryButton({ label: '继续', onTap: () => {
          this.permissionsSeen = true
          this.askLocation()
        } })
        SecondaryButton({ label: '跳过', onTap: () => {
          this.permissionsSeen = true
          this.push('pack', 7)
        } })
      }
      .padding({ left: 16, right: 16, bottom: 34 })
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  packPage() {
    Column() {
      this.top(7)
      this.prompt('创建第一个知识包', '先把一份教材或课件交进来')
      Blank().layoutWeight(1)
      Column({ space: 16 }) {
        Text('拍').width(64).height(64).textAlign(TextAlign.Center).backgroundColor('#F1F4F9').borderRadius(32)
        Text('拍摄教材或课件').fontSize(16).fontFamily('NotoSansSC-SemiBold')
      }
      .width(358)
      .height(190)
      .justifyContent(FlexAlign.Center)
      Blank().layoutWeight(1)
      Text('教材不用于训练模型').fontSize(13).fontColor('#7C848F')
      Column({ space: 16 }) {
        PrimaryButton({ label: '拍照', onTap: () => { this.askCamera() } })
        SecondaryButton({ label: '更多导入方式', onTap: () => { this.push('more', 7) } })
      }
      .padding({ left: 16, right: 16, bottom: 34, top: 12 })
    }
    .width('100%')
    .height('100%')
    .alignItems(HorizontalAlign.Center)
  }

  @Builder
  morePage() {
    Column() {
      this.top(7)
      this.prompt('用其他方式导入', '把教材或课件交进来')
      Column({ space: 16 }) {
        Column({ space: 8 }) {
          Text('传').width(64).height(64).textAlign(TextAlign.Center).backgroundColor('#F1F4F9').borderRadius(32)
          Text('上传文件').fontSize(16).fontFamily('NotoSansSC-SemiBold')
          Text('PDF、PPT、Word、图片或音频').fontSize(13).fontColor('#7C848F')
        }
        .width(358)
        .height(190)
        .justifyContent(FlexAlign.Center)
        .onClick(() => { this.pickFile() })
        Text('或').fontSize(13).fontColor('#7C848F')
        Row({ space: 12 }) {
          Text('抖音').layoutWeight(1).height(50).textAlign(TextAlign.Center).fontSize(16)
            .fontFamily('NotoSansSC-SemiBold').backgroundColor('#FFFFFF').borderRadius(25)
            .onClick(() => { this.showToast() })
          Text('B站').layoutWeight(1).height(50).textAlign(TextAlign.Center).fontSize(16)
            .fontFamily('NotoSansSC-SemiBold').backgroundColor('#FFFFFF').borderRadius(25)
            .onClick(() => { this.showToast() })
        }
        .width(358)
        Text('粘贴笔记')
          .width(358)
          .height(50)
          .textAlign(TextAlign.Center)
          .fontSize(16)
          .fontFamily('NotoSansSC-SemiBold')
          .backgroundColor('#FFFFFF')
          .borderRadius(25)
          .onClick(() => { this.push('paste', 7) })
      }
      .margin({ top: 24 })
      Blank().layoutWeight(1)
      if (this.errorText.length > 0) {
        Text(this.errorText)
          .fontSize(13)
          .fontColor('#7C848F')
          .margin({ bottom: 8 })
      }
      SecondaryButton({
        label: '我没有材料',
        enabled: !this.busy,
        onTap: () => { this.skipNoMaterial() }
      })
        .margin({ left: 16, right: 16, bottom: 34 })
    }
    .width('100%')
    .height('100%')
    .alignItems(HorizontalAlign.Center)
  }

  @Builder
  pastePage() {
    Column() {
      this.top(7)
      this.prompt('把笔记粘过来', '我会从里面挖掘知识点')
      TextArea({ text: this.notes, placeholder: '' })
        .width(358)
        .height(220)
        .margin({ top: 24 })
        .padding(16)
        .backgroundColor('#FFFFFF')
        .borderRadius(22)
        .border({ width: 1, color: '#E1E8F0' })
        .onChange((value: string) => { this.notes = value })
      Blank().layoutWeight(1)
      PrimaryButton({
        label: '继续',
        enabled: FlowLogic.notesContinueEnabled(this.notes),
        onTap: () => {
          this.submitMaterial('paste_note', 'text/plain', this.notes, '')
        }
      })
        .margin({ left: 16, right: 16, bottom: 34 })
    }
    .width('100%')
    .height('100%')
    .alignItems(HorizontalAlign.Center)
  }

  @Builder
  waitPage() {
    Column() {
      this.top(7)
      this.prompt(FlowLogic.waitTitle, FlowLogic.waitCaption)
      Blank().layoutWeight(1)
      Stack() {
        Image($rawfile('reading_still.png'))
          .width(358)
          .height(626)
          .objectFit(ImageFit.Contain)
        Image($rawfile('reading_2s.gif'))
          .width(358)
          .height(626)
          .objectFit(ImageFit.Contain)
          .opacity(this.waitEntry >= 0 ? 1 : 1)
      }
      .width(358)
      .height(626)
      Text(FlowLogic.waitFootnote)
        .width(358)
        .fontSize(16)
        .fontFamily('NotoSansSC-SemiBold')
        .fontColor('#64748B')
        .textAlign(TextAlign.Center)
        .margin({ top: 20 })
      Stack({ alignContent: Alignment.Start }) {
        Row().width(248).height(8).borderRadius(100).backgroundColor('#E1E8F0')
        Row().width(FlowLogic.waitMeterHold).height(8).borderRadius(100)
          .linearGradient({ angle: 90, colors: [['#6C5CE7', 0], ['#D244EE', 1]] })
      }
      .width(248)
      .height(8)
      .margin({ top: 20 })
      if (this.errorText.length > 0) {
        Text(this.errorText)
          .fontSize(13)
          .fontColor('#7C848F')
          .margin({ top: 12 })
      }
      if (this.waitRetryable) {
        SecondaryButton({
          label: '再试一次',
          enabled: !this.busy,
          onTap: () => { this.retryImport() }
        })
          .width(358)
          .margin({ top: 12 })
      }
      Blank().layoutWeight(1)
    }
    .width('100%')
    .height('100%')
    .alignItems(HorizontalAlign.Center)
    .onAppear(() => { this.startJobPoll() })
    .onDisAppear(() => { this.stopJobPoll() })
  }

  @Builder
  top(stepIndex: number) {
    Row({ space: 12 }) {
      BackButton({ onTap: () => { this.back() } })
      StepProgress({ step: stepIndex })
    }
    .width('100%')
    .padding({ left: 16, right: 16, top: 16 })
  }

  @Builder
  prompt(title: string, subtitle: string) {
    Row({ space: 12 }) {
      Image($rawfile('avatar.png'))
        .width(48)
        .height(48)
        .borderRadius(24)
        .objectFit(ImageFit.Cover)
      Column({ space: 4 }) {
        Text(title)
          .fontSize(22)
          .fontFamily('NotoSansSC-ExtraBold')
          .fontColor('#0F172A')
        if (subtitle.length > 0) {
          Text(subtitle)
            .fontSize(16)
            .fontFamily('NotoSansSC-SemiBold')
            .fontColor('#64748B')
        }
      }
      .layoutWeight(1)
      .alignItems(HorizontalAlign.Start)
    }
    .width('100%')
    .padding({ left: 16, right: 16, top: 16 })
  }

  @Builder
  permLine(title: string, body: string, checked: boolean, toggle: () => void) {
    Row({ space: 12 }) {
      Text(title.indexOf('通讯录') >= 0 ? '人' : '铃')
        .width(40)
        .height(40)
        .textAlign(TextAlign.Center)
        .backgroundColor('#F1F4F9')
        .borderRadius(20)
      Column() {
        Text(title).fontSize(16).fontFamily('NotoSansSC-SemiBold')
        Text(body).fontSize(13).fontColor('#64748B')
      }
      .layoutWeight(1)
      .alignItems(HorizontalAlign.Start)
      PermissionCheck({ checked: checked, onTap: toggle })
    }
    .width('100%')
    .height(100)
    .padding({ left: 16, right: 16 })
  }

  @Builder
  sheetPage() {
    Column() {
      Blank().layoutWeight(1).onClick(() => { this.sheet = '' })
      Column({ space: 8 }) {
        Row() {
          Blank().width(40)
          Text(this.sheet === 'region' ? '选择国家或地区' : '选择语言')
            .layoutWeight(1)
            .fontSize(17)
            .fontFamily('NotoSansSC-SemiBold')
            .textAlign(TextAlign.Center)
          Text('×')
            .width(40)
            .height(40)
            .textAlign(TextAlign.Center)
            .backgroundColor('#F2F5F7')
            .borderRadius(20)
            .onClick(() => { this.sheet = '' })
        }
        if (this.sheet === 'region') {
          this.regionRow('中国大陆', '+86')
          this.regionRow('香港', '+852')
        } else {
          this.languageRow('简体中文', 'zh-Hans')
          this.languageRow('繁体中文', 'zh-Hant')
        }
      }
      .width('100%')
      .padding({ left: 16, right: 16, top: 16, bottom: 28 })
      .backgroundColor('#FFFFFF')
      .borderRadius({ topLeft: 28, topRight: 28 })
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#660F172A')
  }

  @Builder
  regionRow(name: string, code: string) {
    Row() {
      Column() {
        Text(name).fontSize(16).fontFamily('NotoSansSC-SemiBold')
        Text(code).fontSize(13).fontColor('#64748B')
      }
      .alignItems(HorizontalAlign.Start)
      .layoutWeight(1)
      Text(this.region === code ? '●' : '○').fontSize(16)
    }
    .width('100%')
    .padding({ top: 16, bottom: 16 })
    .onClick(() => {
      this.region = code
      this.sheet = ''
    })
  }

  @Builder
  languageRow(label: string, locale: string) {
    Row() {
      Text(label).fontSize(16).fontFamily('NotoSansSC-SemiBold').layoutWeight(1)
      Text(this.locale === locale ? '●' : '○').fontSize(16)
    }
    .width('100%')
    .padding({ top: 16, bottom: 16 })
    .onClick(() => {
      this.locale = locale
      this.sheet = ''
    })
  }

  private push(next: string, stepIndex: number): void {
    this.stack.push(this.route)
    this.stepsStack.push(this.step)
    this.route = next
    this.step = stepIndex
    this.errorText = ''
    this.toast = false
  }

  private back(): void {
    this.toast = false
    if (this.route === 'wait') {
      this.resetWaitSession()
    }
    if (this.stack.length === 0) {
      this.onLeave()
      return
    }
    const previous = this.stack.pop()
    const previousStep = this.stepsStack.pop()
    this.route = previous === undefined ? 'auth' : previous
    this.step = previousStep === undefined ? 1 : previousStep
  }

  private showToast(): void {
    this.toast = true
    setTimeout(() => {
      this.toast = false
    }, 2000)
  }

  private resetWaitSession(): void {
    this.stopJobPoll()
    this.waitStart = 0
    this.waitRetryable = false
    this.jobId = ''
    this.packId = ''
    this.materialId = ''
    this.waitEntry = this.waitEntry + 1
  }

  private openWait(jobRaw: string): void {
    this.resetWaitSession()
    this.jobId = jsonField(jobRaw, 'job_id')
    this.packId = jsonField(jobRaw, 'pack_id')
    this.materialId = jsonField(jobRaw, 'material_id')
    this.errorText = ''
    this.push('wait', 7)
  }

  private stopJobPoll(): void {
    if (this.pollTimer >= 0) {
      clearInterval(this.pollTimer)
      this.pollTimer = -1
    }
  }

  private startJobPoll(): void {
    this.stopJobPoll()
    if (this.jobId.length === 0) {
      return
    }
    const tick = (): void => {
      const current = this.jobId
      if (current.length === 0 || this.route !== 'wait') {
        return
      }
      this.api.getJob(this.bearer, current).then((raw: string) => {
        if (this.route !== 'wait' || this.jobId !== current) {
          return
        }
        const status = statusOf(raw)
        if (status !== 200) {
          this.errorText = errorCode(raw)
          return
        }
        const jobStatus = jsonField(raw, 'status')
        const retryable = jsonField(raw, 'retryable') === 'true'
        this.waitRetryable = FlowLogic.waitAllowsRetry(jobStatus, retryable)
        if (FlowLogic.waitReleasesHome(jobStatus)) {
          this.stopJobPoll()
          this.onHome()
          return
        }
        if (jobStatus === FlowLogic.jobFailedRetryable || jobStatus === FlowLogic.jobFailedTerminal) {
          this.stopJobPoll()
          const code = jsonField(raw, 'failure_code')
          this.errorText = code.length > 0 ? code : jobStatus
          return
        }
        this.errorText = ''
      }).catch(() => {
        if (this.route === 'wait') {
          this.errorText = 'transport'
        }
      })
    }
    tick()
    this.pollTimer = setInterval(tick, FlowLogic.jobPollIntervalMillis)
  }

  private submitMaterial(source: string, contentType: string, body: string, filename: string): void {
    if (this.busy || body.length === 0) {
      return
    }
    this.busy = true
    this.errorText = ''
    this.api.createPack(this.bearer).then((packRaw: string) => {
      if (statusOf(packRaw) < 200 || statusOf(packRaw) >= 300) {
        this.busy = false
        this.errorText = errorCode(packRaw)
        return
      }
      const packId = jsonField(packRaw, 'pack_id')
      this.api.createUploadUrl(this.bearer, packId, source, contentType, body.length, filename).then((uploadRaw: string) => {
        if (statusOf(uploadRaw) < 200 || statusOf(uploadRaw) >= 300) {
          this.busy = false
          this.errorText = errorCode(uploadRaw)
          return
        }
        const uploadUrl = jsonField(uploadRaw, 'upload_url')
        const method = jsonField(uploadRaw, 'upload_method')
        const headers = jsonObject(uploadRaw, 'upload_headers')
        this.api.putMaterial(uploadUrl, method.length > 0 ? method : 'PUT', headers, body).then((putStatus: string) => {
          const code = Number(putStatus)
          if (code < 200 || code >= 300) {
            this.busy = false
            this.errorText = 'VALIDATION_MATERIAL_NOT_UPLOADED'
            return
          }
          const materialId = jsonField(uploadRaw, 'material_id')
          this.api.createImport(FlowLogic.newIdempotencyKey(), this.bearer, packId, materialId).then((jobRaw: string) => {
            this.busy = false
            if (statusOf(jobRaw) >= 200 && statusOf(jobRaw) < 300) {
              this.openWait(jobRaw)
              return
            }
            this.errorText = errorCode(jobRaw)
          }).catch(() => {
            this.busy = false
            this.errorText = 'transport'
          })
        }).catch(() => {
          this.busy = false
          this.errorText = 'transport'
        })
      }).catch(() => {
        this.busy = false
        this.errorText = 'transport'
      })
    }).catch(() => {
      this.busy = false
      this.errorText = 'transport'
    })
  }

  private retryImport(): void {
    if (this.busy || this.packId.length === 0 || this.materialId.length === 0) {
      return
    }
    this.busy = true
    this.errorText = ''
    this.api.createImport(FlowLogic.newIdempotencyKey(), this.bearer, this.packId, this.materialId).then((raw: string) => {
      this.busy = false
      if (statusOf(raw) >= 200 && statusOf(raw) < 300) {
        this.jobId = jsonField(raw, 'job_id')
        this.packId = jsonField(raw, 'pack_id')
        this.materialId = jsonField(raw, 'material_id')
        this.waitRetryable = false
        this.errorText = ''
        this.waitEntry = this.waitEntry + 1
        this.startJobPoll()
        return
      }
      this.errorText = errorCode(raw)
    }).catch(() => {
      this.busy = false
      this.errorText = 'transport'
    })
  }

  private pickFile(): void {
    const documentPicker = new picker.DocumentViewPicker()
    documentPicker.select(new picker.DocumentSelectOptions()).then((uris: Array<string>) => {
      if (uris.length === 0) {
        return
      }
      const file = fileIo.openSync(uris[0], fileIo.OpenMode.READ_ONLY)
      const stat = fileIo.statSync(file.fd)
      const buffer = new ArrayBuffer(stat.size)
      fileIo.readSync(file.fd, buffer)
      fileIo.closeSync(file)
      const view = new Uint8Array(buffer)
      let text = ''
      for (let i = 0; i < view.length; i++) {
        text += String.fromCharCode(view[i])
      }
      this.submitMaterial('file', 'application/octet-stream', text, '')
    }).catch(() => {
      this.errorText = 'transport'
    })
  }

  private sendCode(pushOtp: boolean): void {
    if (this.busy) {
      return
    }
    this.busy = true
    this.errorText = ''
    this.api.send(FlowLogic.newIdempotencyKey(), this.region, this.phone, this.mode, this.locale).then((raw: string) => {
      this.busy = false
      const status = statusOf(raw)
      if (status >= 200 && status < 300) {
        this.hint = jsonField(raw, 'destination_hint')
        this.code = ''
        if (pushOtp) {
          this.push('otp', this.step)
        }
        return
      }
      const code = errorCode(raw)
      if (code === 'AUTH_ACCOUNT_NOT_FOUND') {
        this.push('noAccount', this.step)
      } else if (code === 'AUTH_ACCOUNT_EXISTS') {
        this.mode = 'login'
      } else {
        this.errorText = code
      }
    }).catch(() => {
      this.busy = false
      this.errorText = 'transport'
    })
  }

  private verify(): void {
    if (!FlowLogic.otpContinueEnabled(this.code) || this.busy) {
      return
    }
    this.busy = true
    this.api.verify(FlowLogic.newIdempotencyKey(), this.region, this.phone, this.mode, this.code, this.locale).then((raw: string) => {
      this.busy = false
      const status = statusOf(raw)
      if (status >= 200 && status < 300) {
        this.bearer = jsonField(raw, 'access_token')
        const next = jsonField(raw, 'next_step')
        if (next === FlowLogic.nextOnboarding) {
          this.api.onboarding(this.bearer).then((stateRaw: string) => {
            const stepId = jsonField(stateRaw, 'current_step_id')
            const place = FlowLogic.placeForOnboarding(stepId)
            this.stack = []
            this.stepsStack = []
            this.route = place.route
            this.step = place.step
          }).catch(() => {
            this.route = 'nickname'
            this.step = 1
          })
        } else {
          this.follow(next)
        }
        return
      }
      const code = errorCode(raw)
      if (code === 'AUTH_OTP_INVALID' || code === 'AUTH_OTP_EXPIRED' || code === 'AUTH_OTP_MAX_ATTEMPTS') {
        this.code = ''
      }
      this.errorText = code
    }).catch(() => {
      this.busy = false
      this.errorText = 'transport'
    })
  }

  private writeStep(stepId: string, nickname: string, answer: string): void {
    if (this.busy) {
      return
    }
    this.busy = true
    this.api.writeOnboarding(this.bearer, stepId, nickname, answer).then((raw: string) => {
      this.busy = false
      const status = statusOf(raw)
      if (status >= 200 && status < 300) {
        if (stepId === 'daily_time') {
          this.follow(jsonField(raw, 'next_step'))
        } else if (stepId === 'nickname') {
          this.selected = ''
          this.push('question', 2)
        } else {
          this.selected = ''
          this.push('question', this.step + 1)
        }
        return
      }
      this.errorText = errorCode(raw)
    }).catch(() => {
      this.busy = false
      this.errorText = 'transport'
    })
  }

  private pick(label: string): void {
    this.selected = label
    const current: QuestionStep = FlowLogic.question(this.step)
    if (current.finishLabel.length === 0) {
      this.writeStep(current.stepId, '', label)
    }
  }

  private askThen(next: string): void {
    const names: Array<Permissions> = []
    if (this.contactsOn) {
      names.push('ohos.permission.READ_CONTACTS')
    }
    this.requestThen(names, next, 7)
  }

  private askLocation(): void {
    this.requestThen(['ohos.permission.APPROXIMATELY_LOCATION', 'ohos.permission.LOCATION'], 'pack', 7)
  }

  private askCamera(): void {
    try {
      const context = getContext(this) as common.UIAbilityContext
      abilityAccessCtrl.createAtManager().requestPermissionsFromUser(context, ['ohos.permission.CAMERA']).then(() => {
      }).catch(() => {
      })
    } catch (_err) {
    }
  }

  private requestThen(names: Array<Permissions>, next: string, stepIndex: number): void {
    if (names.length === 0) {
      this.push(next, stepIndex)
      return
    }
    try {
      const context = getContext(this) as common.UIAbilityContext
      abilityAccessCtrl.createAtManager().requestPermissionsFromUser(context, names).then(() => {
        this.push(next, stepIndex)
      }).catch(() => {
        this.push(next, stepIndex)
      })
    } catch (_err) {
      this.push(next, stepIndex)
    }
  }

  private skipNoMaterial(): void {
    if (this.busy) {
      return
    }
    this.busy = true
    this.errorText = ''
    this.api.skipFirstPack(FlowLogic.newIdempotencyKey(), this.bearer).then((raw: string) => {
      const status = statusOf(raw)
      if (status === 200) {
        this.applySkip(jsonField(raw, 'next_step'))
        return
      }
      this.busy = false
      this.errorText = errorCode(raw)
    }).catch(() => {
      this.busy = false
      this.errorText = 'transport'
    })
  }

  private applySkip(nextStep: string): void {
    if (nextStep === FlowLogic.nextOnboarding) {
      this.api.onboarding(this.bearer).then((raw: string) => {
        this.busy = false
        const status = statusOf(raw)
        if (status >= 200 && status < 300) {
          const place = FlowLogic.placeForOnboarding(jsonField(raw, 'current_step_id'))
          this.stack = []
          this.stepsStack = []
          this.route = place.route
          this.step = place.step
          return
        }
        this.errorText = errorCode(raw)
      }).catch(() => {
        this.busy = false
        this.errorText = 'transport'
      })
      return
    }
    this.busy = false
    this.follow(nextStep)
  }

  private follow(nextStep: string): void {
    const next = FlowLogic.placeAfterSession(nextStep, this.permissionsSeen)
    if (next === 'home') {
      this.onHome()
      return
    }
    this.stack = []
    this.stepsStack = []
    this.route = next
    this.step = next === 'nickname' ? 1 : 7
    if (next === 'nickname') {
      this.step = 1
    }
  }
}

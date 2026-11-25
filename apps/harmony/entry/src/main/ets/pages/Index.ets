import { getHealth } from '../generated/QlutchClient'
import { API_BASE_URL } from '../config/ApiConfig'
import { font } from '@kit.ArkUI'
import { SplashView } from '../splash/SplashView'
import { WelcomeView } from '../welcome/WelcomeView'
import { FlowHost, NavBus } from '../flow/FlowHost'
import { FlowLogic } from '../flow/FlowLogic'

@Entry
@Component
struct Index {
  @State current: number = 0
  @State health: string = '尚未检查'
  @State phase: string = 'splash'
  @State purpose: string = FlowLogic.purposeRegister
  private fontsReady: boolean = false

  aboutToAppear(): void {
    if (!this.fontsReady) {
      font.registerFont({ familyName: 'NotoSansSC-Regular', familySrc: $rawfile('fonts/noto_sans_sc_regular.ttf') })
      font.registerFont({ familyName: 'NotoSansSC-Medium', familySrc: $rawfile('fonts/noto_sans_sc_medium.ttf') })
      font.registerFont({ familyName: 'NotoSansSC-SemiBold', familySrc: $rawfile('fonts/noto_sans_sc_semibold.ttf') })
      font.registerFont({ familyName: 'NotoSansSC-ExtraBold', familySrc: $rawfile('fonts/noto_sans_sc_extrabold.ttf') })
      this.fontsReady = true
    }
    getHealth(API_BASE_URL).then((result) => {
      this.health = `${result.statusCode} ${result.body.status}`
    }).catch(() => {
      this.health = '不可达'
    })
  }

  onBackPress(): boolean {
    if (this.phase === 'splash') {
      this.phase = 'welcome'
      return true
    }
    if (this.phase === 'flow') {
      return NavBus.back()
    }
    return false
  }

  @Builder
  pageBody(title: string) {
    Column() {
      Text(title)
        .fontSize(32)
      Text('占位')
        .fontSize(16)
        .fontColor('#6b645c')
      Text(`指向 ${API_BASE_URL}`)
        .fontSize(14)
        .fontFamily('monospace')
      Text(`探针 ${this.health}`)
        .fontSize(14)
        .fontFamily('monospace')
      Text('这里没有进度，也没有业务数据。')
        .fontSize(14)
        .fontColor('#6b645c')
    }
    .width('100%')
    .height('100%')
    .alignItems(HorizontalAlign.Start)
    .padding(24)
  }

  @Builder
  tabLabel(title: string, index: number) {
    Text(title)
      .fontSize(12)
      .fontColor(this.current === index ? '#145c4a' : '#6b645c')
      .width('100%')
      .height(48)
      .textAlign(TextAlign.Center)
  }

  @Builder
  shell() {
    Tabs({ barPosition: BarPosition.End, index: this.current }) {
      TabContent() {
        this.pageBody('首页')
      }
      .tabBar(this.tabLabel('首页', 0))

      TabContent() {
        this.pageBody('做题')
      }
      .tabBar(this.tabLabel('做题', 1))

      TabContent() {
        this.pageBody('PK')
      }
      .tabBar(this.tabLabel('PK', 2))

      TabContent() {
        this.pageBody('好友')
      }
      .tabBar(this.tabLabel('好友', 3))

      TabContent() {
        this.pageBody('我的')
      }
      .tabBar(this.tabLabel('我的', 4))
    }
    .onChange((index: number) => {
      this.current = index
    })
  }

  build() {
    if (this.phase === 'splash') {
      SplashView({ onDone: () => { this.phase = 'welcome' } })
    } else if (this.phase === 'welcome') {
      WelcomeView({
        onStart: () => {
          this.purpose = FlowLogic.purposeRegister
          this.phase = 'flow'
        },
        onExisting: () => {
          this.purpose = FlowLogic.purposeLogin
          this.phase = 'flow'
        }
      })
    } else if (this.phase === 'flow') {
      FlowHost({
        purpose: this.purpose,
        onHome: () => { this.phase = 'home' },
        onLeave: () => { this.phase = 'welcome' }
      })
    } else {
      this.shell()
    }
  }
}

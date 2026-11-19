import { getHealth } from '../generated/QlutchClient'
import { API_BASE_URL } from '../config/ApiConfig'

@Entry
@Component
struct Index {
  @State current: number = 0
  @State health: string = '尚未检查'

  aboutToAppear(): void {
    getHealth(API_BASE_URL).then((result) => {
      this.health = `${result.statusCode} ${result.body.status}`
    }).catch(() => {
      this.health = '不可达'
    })
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

  build() {
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
}

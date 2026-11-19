package dev.qlutch.app

import android.os.Bundle
import android.os.Handler
import android.os.Looper
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import dev.qlutch.contract.QlutchClient

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MaterialTheme {
                Shell()
            }
        }
    }
}

private data class Tab(val title: String)

private val tabs = listOf(
    Tab("首页"),
    Tab("做题"),
    Tab("PK"),
    Tab("好友"),
    Tab("我的"),
)

@Composable
private fun Shell() {
    val selected = remember { mutableStateOf(0) }
    val health = remember { mutableStateOf("尚未检查") }
    DisposableEffect(Unit) {
        val worker = Thread {
            val text = try {
                val result = QlutchClient(ApiConfig.BASE_URL).getHealth()
                "${result.statusCode} ${result.body.status}"
            } catch (_: Exception) {
                "不可达"
            }
            Handler(Looper.getMainLooper()).post { health.value = text }
        }
        worker.start()
        onDispose { worker.interrupt() }
    }
    Scaffold(
        bottomBar = {
            NavigationBar {
                NavigationBarItem(
                    selected = selected.value == 0,
                    onClick = { selected.value = 0 },
                    icon = { Text("首") },
                    label = { Text("首页") },
                )
                NavigationBarItem(
                    selected = selected.value == 1,
                    onClick = { selected.value = 1 },
                    icon = { Text("做") },
                    label = { Text("做题") },
                )
                NavigationBarItem(
                    selected = selected.value == 2,
                    onClick = { selected.value = 2 },
                    icon = { Text("P") },
                    label = { Text("PK") },
                )
                NavigationBarItem(
                    selected = selected.value == 3,
                    onClick = { selected.value = 3 },
                    icon = { Text("友") },
                    label = { Text("好友") },
                )
                NavigationBarItem(
                    selected = selected.value == 4,
                    onClick = { selected.value = 4 },
                    icon = { Text("我") },
                    label = { Text("我的") },
                )
            }
        },
    ) { padding ->
        val tab = tabs[selected.value]
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(24.dp),
        ) {
            Text(tab.title, style = MaterialTheme.typography.headlineLarge)
            Text("占位")
            Text("指向 ${ApiConfig.BASE_URL}")
            Text("探针 ${health.value}")
            Text("这里没有进度，也没有业务数据。")
        }
    }
}

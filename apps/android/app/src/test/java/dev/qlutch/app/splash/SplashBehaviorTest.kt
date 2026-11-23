package dev.qlutch.app.splash

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File
import java.security.MessageDigest
import kotlin.math.roundToInt

class SplashBehaviorTest {
    @Test
    fun animatorOrTransitionScaleOffSkipsMotion() {
        assertTrue(isReduceMotion(animatorScaleOff = true, transitionScaleOff = false, platformReduceMotion = false))
        assertTrue(isReduceMotion(animatorScaleOff = false, transitionScaleOff = true, platformReduceMotion = false))
        assertTrue(isReduceMotion(animatorScaleOff = false, transitionScaleOff = false, platformReduceMotion = true))
        assertFalse(isReduceMotion(animatorScaleOff = false, transitionScaleOff = false, platformReduceMotion = false))
    }

    @Test
    fun scaleOffOnlyWhenTheStoredValueIsZero() {
        assertFalse(scaleIsOff(null))
        assertFalse(scaleIsOff(""))
        assertFalse(scaleIsOff("1"))
        assertFalse(scaleIsOff("1.0"))
        assertFalse(scaleIsOff("0.5"))
        assertTrue(scaleIsOff("0"))
        assertTrue(scaleIsOff("0.0"))
        assertTrue(scaleIsOff("0.0f"))
        assertTrue(scaleIsOff("0,0"))
    }

    @Test
    fun welcomeFrameIsTwoThirdsWideAndNineBySixteen() {
        val (width, height) = SplashLayout.frameSize(1080f)
        assertEquals(720f, width, 0.001f)
        assertEquals(1280f, height, 0.001f)
        assertEquals(width * SplashLayout.HEIGHT_OVER_WIDTH, height, 0f)
        assertTrue(width < 1080f)
        assertTrue(height < 1920f)

        val (narrow, narrowHeight) = SplashLayout.frameSize(393f)
        assertEquals(262f, narrow, 0.001f)
        assertEquals(262f * 16f / 9f, narrowHeight, 0.001f)
        assertEquals(SplashLayout.WIDTH_FRACTION, 2f / 3f, 0f)
    }

    @Test
    fun containScaleFitsTheClipWithoutCroppingOrStretching() {
        val (matchedX, matchedY) = containScale(720, 1280, 1080, 1920)
        assertEquals(1f, matchedX, 0.001f)
        assertEquals(1f, matchedY, 0.001f)

        val (wideX, wideY) = containScale(640, 640, 1080, 1920)
        assertEquals((1080f / 1920f) / (640f / 640f), wideX, 0.001f)
        assertEquals(1f, wideY, 0.001f)
        assertTrue(wideX < 1f)

        val (tallX, tallY) = containScale(360, 800, 1080, 1920)
        assertEquals(1f, tallX, 0.001f)
        assertTrue(tallY < 1f)

        assertEquals(1f to 1f, containScale(0, 100, 1080, 1920))
    }

    @Test
    fun staticHoldStaysInsideTheEntranceBudget() {
        assertTrue(SplashTiming.STATIC_HOLD_MILLIS in 1_200L..2_000L)
        assertEquals(1_600L, SplashTiming.STATIC_HOLD_MILLIS)
    }

    @Test
    fun naturalCompletionHoldsTheLastFrameForOneSecondThenLeavesOnce() {
        assertEquals(1_000L, SplashTiming.POST_COMPLETION_HOLD_MILLIS)
        val clock = HoldClock()
        val hold = SplashCompletionHold(schedule = clock::schedule)
        var leaves = 0
        hold.onNaturalCompletion { leaves += 1 }
        hold.onNaturalCompletion { leaves += 1 }
        assertEquals(listOf(SplashTiming.POST_COMPLETION_HOLD_MILLIS), clock.delays)
        clock.advance(SplashTiming.POST_COMPLETION_HOLD_MILLIS - 1)
        assertEquals(0, leaves)
        clock.advance(1)
        assertEquals(1, leaves)
        clock.advance(SplashTiming.POST_COMPLETION_HOLD_MILLIS)
        hold.onNaturalCompletion { leaves += 1 }
        assertEquals(1, leaves)
        assertEquals(1, clock.delays.size)
    }

    @Test
    fun skipDuringThePostCompletionHoldLeavesImmediately() {
        val clock = HoldClock()
        val hold = SplashCompletionHold(schedule = clock::schedule)
        var leaves = 0
        hold.onNaturalCompletion { leaves += 1 }
        clock.advance(400)
        assertEquals(0, leaves)
        hold.onSkip { leaves += 1 }
        assertEquals(1, leaves)
        clock.advance(5_000)
        assertEquals(1, leaves)
        assertEquals(listOf(SplashTiming.POST_COMPLETION_HOLD_MILLIS), clock.delays)

        val playing = HoldClock()
        val duringPlayback = SplashCompletionHold(schedule = playing::schedule)
        var skipped = 0
        duringPlayback.onSkip { skipped += 1 }
        assertEquals(1, skipped)
        assertTrue(playing.delays.isEmpty())
        duringPlayback.onNaturalCompletion { skipped += 1 }
        playing.advance(SplashTiming.POST_COMPLETION_HOLD_MILLIS)
        assertEquals(1, skipped)

        val released = HoldClock()
        val held = SplashCompletionHold(schedule = released::schedule)
        var releasedLeaves = 0
        held.onNaturalCompletion { releasedLeaves += 1 }
        held.cancelPending()
        released.advance(SplashTiming.POST_COMPLETION_HOLD_MILLIS)
        assertEquals(0, releasedLeaves)
    }

    @Test
    fun playbackUsesThePostCompletionHoldAndSkipStaysImmediate() {
        val timing = moduleFile("src/main/java/dev/qlutch/app/splash/SplashTiming.kt").readText()
        val video = moduleFile("src/main/java/dev/qlutch/app/splash/SplashVideoView.kt").readText()
        val playback = moduleFile("src/main/java/dev/qlutch/app/splash/ColdStartVideo.kt").readText()
        val activity = moduleFile("src/main/java/dev/qlutch/app/splash/SplashActivity.kt").readText()
        val still = moduleFile("src/main/java/dev/qlutch/app/splash/StaticWelcomeStill.kt").readText()

        assertTrue(timing.contains("const val POST_COMPLETION_HOLD_MILLIS = 1_000L"))
        assertTrue(timing.contains("holdMillis: Long = SplashTiming.POST_COMPLETION_HOLD_MILLIS"))
        assertTrue(timing.contains("const val STATIC_HOLD_MILLIS = 1_600L"))

        val completion = video.substringAfter("setOnCompletionListener").substringBefore("setOnErrorListener")
        assertTrue(completion.contains("finished = true"))
        assertTrue(completion.contains("onNaturalCompletion"))
        assertFalse(completion.contains("start()"))
        assertFalse(completion.contains("isLooping = true"))
        val skip = video.substringAfter("fun skip()").substringBefore("fun startPlayback()")
        assertTrue(skip.contains("finished = true"))
        assertTrue(skip.contains("onSkip"))
        assertFalse(skip.contains("postDelayed"))
        assertFalse(skip.contains("1_000L"))
        assertTrue(video.contains("player.isLooping = false"))
        assertTrue(video.contains("if (released || !prepared || finished || surface == null) return"))
        assertTrue(video.contains("completionHold.cancelPending()"))
        assertFalse(video.contains("isLooping = true"))

        assertTrue(playback.contains("onTap = { videoView.value?.skip() ?: onFinished() }"))
        assertTrue(playback.contains("video.onFinished = onFinished"))
        assertFalse(playback.contains("POST_COMPLETION_HOLD_MILLIS"))
        assertFalse(playback.contains("delay("))

        assertTrue(
            activity.contains(
                "override fun handleOnBackPressed() {\n                    openNext()\n                }",
            ),
        )
        assertTrue(activity.contains("WelcomeActivity::class.java"))
        assertFalse(activity.contains("MainActivity::class.java"))
        assertFalse(activity.contains("POST_COMPLETION_HOLD_MILLIS"))
        assertFalse(activity.contains("postDelayed"))
        assertFalse(activity.contains("delay("))

        assertTrue(still.contains("delay(SplashTiming.STATIC_HOLD_MILLIS)"))
        assertFalse(still.contains("POST_COMPLETION_HOLD_MILLIS"))
        assertFalse(still.contains("SplashVideoView"))
    }

    @Test
    fun brandCopyIsWordmarkPeriodAndSloganOnly() {
        val strings = moduleFile("src/main/res/values/strings.xml").readText()
        assertTrue(strings.contains("<string name=\"splash_wordmark\">Qlutch</string>"))
        assertTrue(strings.contains("<string name=\"splash_wordmark_period\">\u3002</string>"))
        assertTrue(strings.contains("<string name=\"splash_slogan\">让记知识，更轻松有趣</string>"))
        assertFalse(strings.contains("变得轻松有趣"))
        assertFalse(strings.contains("把课件或教材"))
        assertFalse(strings.contains("好玩"))
        assertFalse(strings.contains("Gizmo"))

        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"canvas\">#F6F8FF</color>"))
        assertTrue(colors.contains("<color name=\"ink\">#20244A</color>"))
        assertTrue(colors.contains("<color name=\"spark\">#FF6B4A</color>"))
    }

    @Test
    fun splashSourcesDoNotInventPositioningOrIp() {
        val source = moduleFile("src/main/java/dev/qlutch/app/splash")
            .walk()
            .filter { it.extension == "kt" }
            .joinToString("\n") { it.readText() }
        assertFalse(source.contains("把课件或教材上传给AI"))
        assertFalse(source.contains("Gizmo"))
        assertFalse(source.contains("美西螈"))
        assertFalse(source.contains("好玩"))
    }

    @Test
    fun firstFrameStillCoversThePlateUntilTheVideoPaints() {
        val file = moduleFile("src/main/res/drawable-nodpi/splash_first_frame.png")
        val bytes = file.readBytes()
        assertEquals(589309, bytes.size)
        assertEquals(
            "b3a459008a1e233c34bd695c4849cf4d962c2e03060154351f4e071b8a7f373d",
            sha256(bytes),
        )
        val image = decodePng(bytes)
        assertEquals(1080, image.width)
        assertEquals(1920, image.height)
        assertEquals(1080f / 1920f, image.width.toFloat() / image.height.toFloat(), 0.0001f)
        val corner = Triple(image.red(0, 0), image.green(0, 0), image.blue(0, 0))
        val body = Triple(image.red(540, 1400), image.green(540, 1400), image.blue(540, 1400))
        assertTrue(corner != Triple(255, 255, 255))
        assertTrue(body != corner)

        val cover = SplashPosterCover()
        assertTrue(cover.showingStill)
        cover.onPlayerPrepared()
        assertTrue("prepared player is still a blank plate", cover.showingStill)
        cover.onSurfaceAttached()
        assertTrue("a new surface can still be empty", cover.showingStill)
        cover.onFirstFrameRendered()
        assertFalse(cover.showingStill)
        cover.onPlayerPrepared()
        cover.onSurfaceAttached()
        cover.onFirstFrameRendered()
        assertFalse(cover.showingStill)

        val video = moduleFile("src/main/java/dev/qlutch/app/splash/SplashVideoView.kt").readText()
        val still = moduleFile("src/main/java/dev/qlutch/app/splash/StaticWelcomeStill.kt").readText()
        val themes = moduleFile("src/main/res/values/themes.xml").readText()
        val activity = moduleFile("src/main/java/dev/qlutch/app/splash/SplashActivity.kt").readText()

        assertTrue(video.contains("setImageResource(R.drawable.splash_first_frame)"))
        assertTrue(video.contains("scaleType = ImageView.ScaleType.FIT_CENTER"))
        assertTrue(video.contains("posterCover.onPlayerPrepared()"))
        assertTrue(video.contains("posterCover.onSurfaceAttached()"))
        assertTrue(video.contains("posterCover.onFirstFrameRendered()"))
        assertTrue(video.contains("if (posterCover.showingStill) View.VISIBLE else View.INVISIBLE"))
        assertTrue(video.indexOf("addView(\n            video") < video.indexOf("addView(\n            poster"))
        assertTrue(video.indexOf("addView(\n            poster") < video.indexOf("addView(\n            wordmark"))
        assertTrue(video.indexOf("addView(\n            wordmark") < video.indexOf("addView(\n            slogan"))
        assertTrue(video.contains("layoutSplashSlogan"))
        assertTrue(video.contains("wordmarkInFrame"))

        val prepared = video.substringAfter("setOnPreparedListener").substringBefore("setOnVideoSizeChangedListener")
        assertTrue(prepared.contains("onPlaybackPrepared?.invoke()"))
        assertFalse(prepared.contains("onFirstFrameRendered"))
        assertFalse(prepared.contains("INVISIBLE"))
        assertFalse(prepared.contains("splash_first_frame"))

        val surface = video.substringAfter("fun onSurfaceTextureAvailable").substringBefore("fun onSurfaceTextureSizeChanged")
        assertTrue(surface.contains("onSurfaceReady?.invoke()"))
        assertFalse(surface.contains("onFirstFrameRendered"))

        val updated = video.substringAfter("fun onSurfaceTextureUpdated").substringBefore("fun onSizeChanged")
        assertTrue(updated.contains("notifyFirstFrameRendered()"))
        val notify = video.substringAfter("fun notifyFirstFrameRendered()").substringBefore("fun applyContain()")
        assertTrue(notify.contains("if (released || firstFrameNotified || !prepared) return"))
        assertTrue(notify.contains("renderingStarted || textureUpdatesAfterPrepare >= 2"))
        assertTrue(notify.contains("onFirstFrameRendered?.invoke()"))
        assertFalse(notify.contains("start()"))
        val info = video.substringAfter("setOnInfoListener").substringBefore("fun skip()")
        assertTrue(info.contains("MEDIA_INFO_VIDEO_RENDERING_START"))
        assertTrue(info.contains("renderingStarted = true"))
        assertFalse(info.contains("onFirstFrameRendered"))

        val skip = video.substringAfter("fun skip()").substringBefore("fun startPlayback()")
        assertTrue(skip.contains("onSkip"))
        assertFalse(skip.contains("showingStill"))
        assertFalse(skip.contains("splash_first_frame"))
        assertFalse(skip.contains("postDelayed"))

        assertTrue(still.contains("painter = painterResource(R.drawable.splash_first_frame)"))
        assertTrue(still.contains("ContentScale.Fit"))
        assertTrue(still.contains("G4Wordmark()"))
        assertTrue(still.contains("SplashSloganHost"))
        assertTrue(still.contains("delay(SplashTiming.STATIC_HOLD_MILLIS)"))
        assertFalse(still.contains("MediaMetadataRetriever"))
        assertFalse(still.contains("SplashVideoView"))
        assertFalse(still.contains("POST_COMPLETION_HOLD_MILLIS"))

        assertTrue(themes.contains("android:windowBackground\">@drawable/splash_window_background"))
        assertTrue(themes.contains("android:windowSplashScreenBackground\" tools:targetApi=\"31\">@color/splash_background"))
        assertTrue(themes.contains("android:windowSplashScreenAnimatedIcon\" tools:targetApi=\"31\">@drawable/splash_blank_icon"))
        assertTrue(themes.contains("android:windowSplashScreenAnimationDuration\" tools:targetApi=\"31\">0"))
        assertFalse(themes.contains("windowSplashScreenIconBackgroundColor"))
        assertFalse(themes.contains("splash_system_icon"))
        assertFalse(themes.contains("icon_preferred"))
        assertFalse(themes.contains("splash_empty_icon"))
        assertFalse(themes.contains("splash_first_frame"))
        assertFalse(themes.contains("#FFFFFF"))
        assertTrue(activity.contains("StaticWelcomeStill(onContinue = ::openNext)"))
        assertTrue(activity.contains("ColdStartVideo(onFinished = ::openNext)"))
        assertTrue(
            activity.contains(
                "override fun handleOnBackPressed() {\n                    openNext()\n                }",
            ),
        )
    }

    @Test
    fun packedWelcomeClipBytesAreTheCleanEdgeV4File() {
        val file = moduleFile("src/main/res/raw/splash.mp4")
        val bytes = file.readBytes()
        assertEquals(768507, bytes.size)
        assertEquals("b6d02b58cbf8ff491ca063112e24c6a38f406f80516d03c762ad97abf7e3b610", sha256(bytes))
        assertTrue(bytes.containsSequence("vide".encodeToByteArray()))
        assertFalse(bytes.containsSequence("soun".encodeToByteArray()))
    }

    @Test
    fun wordmarkStaysBelowTheLandingShadowAndDoesNotFollowTheJump() {
        assertEquals(1462f, SplashLayout.LANDING_SHADOW_BOTTOM_Y, 0f)
        assertTrue(SplashLayout.LANDING_SHADOW_BOTTOM_Y > 1454f)
        assertTrue(SplashLayout.WORDMARK_GAP_AT_375 in 20f..32f)

        val place = SplashLayout.placement(1080f, 2340f)
        assertEquals(720f, place.frameWidth, 0.001f)
        assertEquals(1280f, place.frameHeight, 0.001f)
        assertEquals(180f, place.frameLeft, 0.001f)
        assertEquals(530f, place.frameTop, 0.001f)
        assertEquals(396f, place.wordmarkWidth, 0.001f)
        assertEquals(396f * 238f / 941f, place.wordmarkHeight, 0.001f)
        assertEquals(342f, place.wordmarkLeft, 0.001f)
        assertEquals(1580f, place.wordmarkTop, 0.5f)
        assertEquals(26f * 1080f / 375f, place.wordmarkGap, 0.01f)
        assertEquals(place.shadowBottom + place.wordmarkGap, place.wordmarkTop, 0.001f)
        assertTrue(place.wordmarkTop > place.shadowBottom)
        assertTrue(place.wordmarkWidth < place.frameWidth)
        assertTrue(place.wordmarkWidth > place.frameWidth * 0.45f)
        assertTrue(place.frameWidth < 1080f)

        val local = SplashLayout.wordmarkInFrame(place.frameWidth, place.frameHeight)
        assertEquals(place.wordmarkTop, place.frameTop + local.top, 0.001f)
        assertEquals(place.wordmarkLeft, place.frameLeft + local.left, 0.001f)
        assertEquals(place, SplashLayout.placement(1080f, 2340f))

        val design = SplashLayout.placement(375f, 812f)
        assertEquals(26f, design.wordmarkGap, 0.001f)
        assertTrue(design.wordmarkTop > design.shadowBottom)

        listOf(360f to 780f, 393f to 852f, 411f to 891f).forEach { (width, height) ->
            val sized = SplashLayout.placement(width, height)
            assertEquals(width * 2f / 3f, sized.frameWidth, 0.05f)
            assertEquals(sized.frameWidth * 16f / 9f, sized.frameHeight, 0.05f)
            assertEquals(
                SplashLayout.WORDMARK_GAP_AT_375 * (width / SplashLayout.DESIGN_WIDTH),
                sized.wordmarkGap,
                0.01f,
            )
            assertTrue(sized.wordmarkTop > sized.shadowBottom)
            assertTrue(sized.wordmarkWidth < width)
        }
    }

    @Test
    fun packedWordmarkIsPurpleG4WithoutAWhitePlate() {
        val file = moduleFile("src/main/res/drawable-nodpi/splash_g4_wordmark.png")
        val bytes = file.readBytes()
        assertEquals(
            "ac47a0ff712f9af63632ac44c7a2014ab7fe71ca18afd48587450b2dd88cc92b",
            sha256(bytes),
        )
        val image = decodePng(bytes)
        assertEquals(941, image.width)
        assertEquals(238, image.height)
        listOf(0 to 0, image.width - 1 to 0, 0 to image.height - 1, image.width - 1 to image.height - 1)
            .forEach { (x, y) ->
                assertEquals(0, image.alpha(x, y))
            }
        var colored = 0
        for (y in 0 until image.height) {
            for (x in 0 until image.width) {
                val alpha = image.alpha(x, y)
                if (alpha == 0) continue
                colored += 1
                assertEquals(0x6C, image.red(x, y))
                assertEquals(0x5C, image.green(x, y))
                assertEquals(0xE7, image.blue(x, y))
            }
        }
        assertTrue(colored > 0)
    }

    @Test
    fun previewSplashKeepsTheV4BackgroundAndOverlaysTheFixedWordmark() {
        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"splash_background\">#F2F5F7</color>"))
        val themes = moduleFile("src/main/res/values/themes.xml").readText()
        assertTrue(themes.contains("android:windowBackground\">@drawable/splash_window_background"))
        assertTrue(themes.contains("android:windowSplashScreenBackground\" tools:targetApi=\"31\">@color/splash_background"))
        assertFalse(themes.contains("@color/canvas"))

        val splashDir = moduleFile("src/main/java/dev/qlutch/app/splash")
        assertFalse(splashDir.list().orEmpty().any { it.contains("StaticBrand") })
        val shown = listOf(
            "ColdStartVideo.kt",
            "StaticWelcomeStill.kt",
            "SplashActivity.kt",
            "SplashFrame.kt",
            "SplashVideoView.kt",
            "SplashSlogan.kt",
        ).joinToString("\n") { name ->
            File(splashDir, name).readText()
        }
        assertTrue(shown.contains("splash_slogan"))
        assertTrue(shown.contains("splash_brand"))
        assertTrue(shown.contains("sloganInFrame"))
        assertTrue(shown.contains("layoutSplashSlogan"))
        assertFalse(shown.contains("变得轻松有趣"))
        assertFalse(shown.contains("coverScale"))
        assertFalse(shown.contains("setColorFilter"))
        assertFalse(shown.contains("colorFilter"))
        assertFalse(shown.contains("R.color.ink"))
        assertFalse(shown.contains("R.color.spark"))
        assertFalse(shown.contains("#20244A"))
        assertFalse(shown.contains("#FF6B4A"))
        assertFalse(shown.contains("currentPosition"))
        assertTrue(shown.contains("splash_background"))
        assertTrue(shown.contains("StaticWelcomeStill"))
        assertTrue(shown.contains("containScale"))
        assertTrue(shown.contains("setVolume(0f, 0f)"))
        assertTrue(shown.contains("OnBackPressedCallback"))
        assertTrue(shown.contains("ContentScale.Fit"))
        assertTrue(shown.contains("ScaleType.FIT_CENTER"))
        assertTrue(shown.contains("splash_first_frame"))
        assertFalse(shown.contains("getFrameAtIndex"))
        assertFalse(shown.contains("MediaMetadataRetriever"))
        assertTrue(shown.contains("splash_g4_wordmark"))
        assertTrue(shown.contains("LANDING_SHADOW_BOTTOM_Y"))
        assertTrue(shown.contains("wordmarkInFrame"))
        assertTrue(shown.contains("WIDTH_FRACTION = 2f / 3f"))
        assertTrue(shown.contains("Color.TRANSPARENT"))
        assertTrue(shown.contains("imageTintList = null"))
        assertTrue(moduleFile("build.gradle.kts").readText().contains("noCompress += \"mp4\""))

        val still = File(splashDir, "StaticWelcomeStill.kt").readText()
        val playback = File(splashDir, "ColdStartVideo.kt").readText()
        val video = File(splashDir, "SplashVideoView.kt").readText()
        val frame = File(splashDir, "SplashFrame.kt").readText()
        assertTrue(still.contains("G4Wordmark()"))
        assertTrue(still.contains("SplashSloganHost"))
        assertFalse(playback.contains("G4Wordmark"))
        assertTrue(playback.contains("SplashPlaybackView"))
        assertTrue(frame.contains("fun G4Wordmark"))
        assertTrue(frame.contains("fun sloganInFrame"))
        assertTrue(frame.contains("R.drawable.splash_g4_wordmark"))
        assertTrue(video.contains("R.drawable.splash_g4_wordmark"))
        assertTrue(video.contains("wordmarkInFrame"))
        assertTrue(video.contains("layoutSplashSlogan"))
        assertTrue(video.indexOf("addView(\n            video") < video.indexOf("addView(\n            wordmark"))
        assertTrue(video.indexOf("addView(\n            wordmark") < video.indexOf("addView(\n            slogan"))
        val slogan = File(splashDir, "SplashSlogan.kt").readText()
        assertTrue(slogan.contains("R.string.splash_slogan"))
        assertTrue(slogan.contains("R.color.splash_brand"))
        assertTrue(slogan.contains("R.font.noto_sans_sc_medium"))
        assertTrue(slogan.contains("PingFang"))
        assertTrue(slogan.contains("fun layoutSplashSlogan"))
        assertTrue(slogan.contains("class SplashSloganHost"))
    }

    @Test
    fun sloganStaysUnderTheFixedWordmarkAndFitsItsWidth() {
        assertEquals(8f, SplashLayout.WORDMARK_INK_TOP, 0f)
        assertEquals(229f, SplashLayout.WORDMARK_INK_BOTTOM, 0f)
        assertEquals(222f, SplashLayout.WORDMARK_INK_HEIGHT, 0f)
        assertEquals(10f, SplashLayout.SLOGAN_GAP_AT_375, 0f)
        assertTrue(SplashLayout.SLOGAN_GAP_AT_375 in 8f..12f)
        assertTrue(SplashLayout.SLOGAN_SIZE_OVER_INK_HEIGHT in 0.40f..0.45f)
        assertEquals(2f / 3f, SplashLayout.WIDTH_FRACTION, 0f)

        val wordmark = decodePng(moduleFile("src/main/res/drawable-nodpi/splash_g4_wordmark.png").readBytes())
        var inkTop = wordmark.height
        var inkBottom = 0
        for (y in 0 until wordmark.height) {
            for (x in 0 until wordmark.width) {
                if (wordmark.alpha(x, y) == 0) continue
                if (y < inkTop) inkTop = y
                if (y > inkBottom) inkBottom = y
            }
        }
        assertEquals(SplashLayout.WORDMARK_INK_TOP, inkTop.toFloat(), 0f)
        assertEquals(SplashLayout.WORDMARK_INK_BOTTOM, inkBottom.toFloat(), 0f)

        val design = SplashLayout.placement(375f, 812f)
        assertEquals(10f, design.sloganGap, 0.001f)
        assertEquals(design.wordmarkLeft, design.sloganLeft, 0.001f)
        assertEquals(design.wordmarkWidth, design.sloganWidth, 0.001f)
        assertEquals(design.wordmarkTop + design.wordmarkHeight + design.sloganGap, design.sloganTop, 0.001f)
        assertTrue(design.sloganGap / design.wordmarkHeight in 0.20f..0.30f)
        val designInk = design.wordmarkHeight * SplashLayout.WORDMARK_INK_HEIGHT / SplashLayout.WORDMARK_ASPECT_HEIGHT
        assertEquals(SplashLayout.SLOGAN_SIZE_OVER_INK_HEIGHT, design.sloganFontSize / designInk, 0.001f)
        assertTrue(design.sloganTextWidth <= design.sloganWidth + 0.001f)
        assertEquals(
            design.wordmarkLeft + design.wordmarkWidth / 2f,
            design.sloganLeft + design.sloganWidth / 2f,
            0.001f,
        )

        val phone = SplashLayout.placement(1080f, 2340f)
        assertEquals(720f, phone.frameWidth, 0.001f)
        assertEquals(180f, phone.frameLeft, 0.001f)
        assertEquals((1080f - 720f) / 2f, phone.frameLeft, 0.001f)
        assertEquals((2340f - 1280f) / 2f, phone.frameTop, 0.001f)
        assertEquals(10f * 1080f / 375f, phone.sloganGap, 0.01f)
        assertEquals(phone.wordmarkTop + phone.wordmarkHeight + phone.sloganGap, phone.sloganTop, 0.01f)
        assertTrue(phone.sloganTop > phone.wordmarkTop + phone.wordmarkHeight)
        val phoneInk = phone.wordmarkHeight * SplashLayout.WORDMARK_INK_HEIGHT / SplashLayout.WORDMARK_ASPECT_HEIGHT
        assertTrue(phone.sloganFontSize / phoneInk in 0.40f..0.45f)
        assertTrue(phone.sloganTextWidth <= phone.wordmarkWidth + 0.001f)
        assertEquals(phone.sloganFontSize * SplashLayout.SLOGAN_ADVANCE_EMS, phone.sloganTextWidth, 0.001f)
        val local = SplashLayout.sloganInFrame(phone.frameWidth, phone.frameHeight)
        assertEquals(phone.sloganTop, phone.frameTop + local.top, 0.001f)
        assertEquals(phone.sloganLeft, phone.frameLeft + local.left, 0.001f)
        assertEquals(phone, SplashLayout.placement(1080f, 2340f))

        listOf(360f to 780f, 393f to 852f, 411f to 891f).forEach { (width, height) ->
            val sized = SplashLayout.placement(width, height)
            assertEquals(width * 2f / 3f, sized.frameWidth, 0.05f)
            assertEquals(
                SplashLayout.SLOGAN_GAP_AT_375 * (width / SplashLayout.DESIGN_WIDTH),
                sized.sloganGap,
                0.01f,
            )
            assertTrue(sized.sloganGap / sized.wordmarkHeight in 0.20f..0.30f)
            val ink = sized.wordmarkHeight * SplashLayout.WORDMARK_INK_HEIGHT / SplashLayout.WORDMARK_ASPECT_HEIGHT
            assertTrue(sized.sloganFontSize / ink in 0.40f..0.45f)
            assertTrue(sized.sloganTextWidth <= sized.wordmarkWidth + 0.01f)
            assertEquals(sized.wordmarkLeft, sized.sloganLeft, 0.01f)
            assertTrue(sized.sloganTop > sized.shadowBottom)
        }

        assertEquals(10f, fittedSloganFontSize(preferred = 10f, measuredWidth = 8f, maxWidth = 12f), 0f)
        assertEquals(10f, fittedSloganFontSize(preferred = 10f, measuredWidth = 10f, maxWidth = 12f), 0f)
        assertEquals(6f, fittedSloganFontSize(preferred = 12f, measuredWidth = 20f, maxWidth = 10f), 0.001f)

        assertEquals(
            SloganTypefaceSource.PINGFANG_SC_MEDIUM,
            chooseSloganTypefaceSource(pingFangMediumAvailable = true),
        )
        assertEquals(
            SloganTypefaceSource.BUNDLED_NOTO_SANS_SC_MEDIUM,
            chooseSloganTypefaceSource(pingFangMediumAvailable = false),
        )

        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"splash_brand\">#6C5CE7</color>"))
        val font = moduleFile("src/main/res/font/noto_sans_sc_medium.ttf").readBytes()
        assertEquals(4608, font.size)
        assertEquals("25ca7ae4b4919637b77ffdf9756b868ec3b2af20e41cce87b4b49f8541d769e5", sha256(font))
    }

    @Test
    fun launcherIconIsTheClawEdgePreview() {
        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"ic_launcher_background\">#F2F5F7</color>"))
        val adaptive = moduleFile("src/main/res/mipmap-anydpi-v26/ic_launcher.xml").readText()
        val round = moduleFile("src/main/res/mipmap-anydpi-v26/ic_launcher_round.xml").readText()
        assertTrue(adaptive.contains("@color/ic_launcher_background"))
        assertTrue(adaptive.contains("@drawable/ic_launcher_foreground"))
        assertTrue(round.contains("@drawable/ic_launcher_foreground"))
        assertTrue(moduleFileIfExists("src/main/res/drawable/ic_launcher_foreground.xml") == null)

        val legacy = mapOf(
            "mipmap-mdpi" to 48,
            "mipmap-hdpi" to 72,
            "mipmap-xhdpi" to 96,
            "mipmap-xxhdpi" to 144,
            "mipmap-xxxhdpi" to 192,
        )
        val legacyHashes = mapOf(
            "mipmap-mdpi" to "90c708072dbee5cedadc26e014834b4821737680c94d1971a8a03dae5439edff",
            "mipmap-hdpi" to "3c8d49cc73308ab138c12b8ec069c406dcbe782b203b7a9298976aaf574a0d9a",
            "mipmap-xhdpi" to "86521455fa5cec08469de7683db58ca03d124b4f49aeb505e272950fc9869049",
            "mipmap-xxhdpi" to "1d80bcec9a6fae0876e9b534289bf19a87deee7093e442b2f1521f2d5bea11dc",
            "mipmap-xxxhdpi" to "dbd6c1c4adff488b2a3f001a68b5a6acacb0862015e0c099f7202e455c27d577",
        )
        legacy.forEach { (dir, size) ->
            val launcher = moduleFile("src/main/res/$dir/ic_launcher.png").readBytes()
            val roundIcon = moduleFile("src/main/res/$dir/ic_launcher_round.png").readBytes()
            assertEquals(launcherHashes(legacyHashes, dir), sha256(launcher))
            assertEquals(sha256(launcher), sha256(roundIcon))
            val image = decodePng(launcher)
            assertEquals(size, image.width)
            assertEquals(size, image.height)
            assertTrue(image.red(image.width / 2, image.height / 2) < 80)
        }

        val foreground = mapOf(
            "drawable-mdpi" to 108,
            "drawable-hdpi" to 162,
            "drawable-xhdpi" to 216,
            "drawable-xxhdpi" to 324,
            "drawable-xxxhdpi" to 432,
        )
        val foregroundHashes = mapOf(
            "drawable-mdpi" to "d5f03c9bc45f71e02241c0d768384f71443f064cad9c625f4203ac1e4f9e8c55",
            "drawable-hdpi" to "7991350d6bf111465eec180b5d4030eeb75bdd052b8bbd64c5b0cb013d498a58",
            "drawable-xhdpi" to "ae5dd178226c464d9dd6cde56739860cb4c5e78a3ae9756625b08e3a22b251d2",
            "drawable-xxhdpi" to "b8f140cabf9558f6d8289ebdab5f0063b2fc30c293940b4236fbb1a74fdf1e22",
            "drawable-xxxhdpi" to "2612054b18c53498c6febf40df21d9de2be8ab7474c2424e5338ef8f2cbe182a",
        )
        foreground.forEach { (dir, size) ->
            val bytes = moduleFile("src/main/res/$dir/ic_launcher_foreground.png").readBytes()
            assertEquals(foregroundHashes.getValue(dir), sha256(bytes))
            val image = decodePng(bytes)
            assertEquals(size, image.width)
            assertEquals(size, image.height)
            listOf(0 to 0, image.width - 1 to 0, 0 to image.height - 1, image.width - 1 to image.height - 1)
                .forEach { (x, y) ->
                    assertEquals(0xF2, image.red(x, y), 8)
                    assertEquals(0xF5, image.green(x, y), 8)
                    assertEquals(0xF7, image.blue(x, y), 8)
                }
        }
    }

    @Test
    fun systemSplashMatchesTheInAppFirstFrameInsteadOfASmallerCircle() {
        val themes = moduleFile("src/main/res/values/themes.xml").readText()
        assertTrue(themes.contains("android:windowSplashScreenBackground\" tools:targetApi=\"31\">@color/splash_background"))
        assertTrue(themes.contains("android:windowSplashScreenAnimatedIcon\" tools:targetApi=\"31\">@drawable/splash_blank_icon"))
        assertTrue(themes.contains("android:windowSplashScreenAnimationDuration\" tools:targetApi=\"31\">0"))
        assertTrue(themes.contains("android:windowBackground\">@drawable/splash_window_background"))
        assertFalse(themes.contains("windowSplashScreenIconBackgroundColor"))
        assertFalse(themes.contains("windowSplashScreenBrandingImage"))
        assertFalse(themes.contains("splash_system_icon"))
        assertFalse(themes.contains("icon_preferred"))
        assertFalse(themes.contains("splash_empty_icon"))
        assertFalse(themes.contains("splash_first_frame"))
        assertFalse(themes.contains("#FFFFFF"))
        assertTrue(moduleFileIfExists("src/main/res/drawable/splash_empty_icon.xml") == null)
        assertTrue(moduleFileIfExists("src/main/res/drawable-nodpi/splash_system_icon.png") == null)

        val blank = moduleFile("src/main/res/drawable/splash_blank_icon.xml").readText()
        assertTrue(blank.contains("@color/splash_background"))
        assertFalse(blank.contains("splash_first_frame"))
        assertFalse(blank.contains("splash_g4_wordmark"))
        assertFalse(blank.contains("splash_slogan"))

        val window = moduleFile("src/main/res/drawable/splash_window_background.xml").readText()
        assertTrue(window.contains("@color/splash_background"))
        assertFalse(window.contains("splash_first_frame"))
        assertFalse(window.contains("250dp"))
        assertFalse(window.contains("444.444dp"))
        assertFalse(window.contains("splash_g4_wordmark"))

        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"splash_background\">#F2F5F7</color>"))
        assertTrue(colors.contains("<color name=\"splash_brand\">#6C5CE7</color>"))

        val phone = SplashLaunchGeometry.framePx(1080, 2340, 3f)
        assertEquals(720, phone.width)
        assertEquals(1280, phone.height)
        assertEquals(180, phone.left)
        assertEquals(530, phone.top)
        val placed = SplashLayout.placement(1080f, 2340f)
        assertEquals(placed.frameLeft.roundToInt(), phone.left)
        assertEquals(placed.frameTop.roundToInt(), phone.top)
        assertEquals(placed.frameWidth.roundToInt(), phone.width)
        assertEquals(placed.frameHeight.roundToInt(), phone.height)
        val mark = SplashLaunchGeometry.wordmarkInFramePx(phone.width, phone.height)
        assertEquals(placed.wordmarkLeft.roundToInt(), phone.left + mark.left)
        assertEquals(placed.wordmarkTop.roundToInt(), phone.top + mark.top)
        assertEquals(placed.wordmarkWidth.roundToInt(), mark.width)
        assertEquals(placed.wordmarkHeight.roundToInt(), mark.height)
        val slogan = sloganDrawPlacement(
            phone.width.toFloat(),
            phone.height.toFloat(),
            measuredTextWidth = { fontSize -> fontSize * SplashLayout.SLOGAN_ADVANCE_EMS },
            inkInsetPx = { 4f },
        )
        val expectedSlogan = SplashLayout.sloganInFrame(phone.width.toFloat(), phone.height.toFloat())
        assertEquals(expectedSlogan.left, slogan.slotLeft, 0.001f)
        assertEquals(expectedSlogan.top, slogan.slotTop, 0.001f)
        assertEquals(expectedSlogan.width, slogan.slotWidth, 0.001f)
        assertEquals(expectedSlogan.fontSize, slogan.fontSizePx, 0.001f)
        assertEquals(expectedSlogan.top - 4f, slogan.viewTop, 0.001f)
        assertTrue(slogan.slotTop > mark.top + mark.height)

        val short = SplashLaunchGeometry.framePx(1080, 1920, 3f)
        val tall = SplashLaunchGeometry.framePx(1080, 2400, 3f)
        assertEquals(short.width, tall.width)
        assertEquals(short.height, tall.height)
        assertTrue(tall.top > short.top)
        assertTrue(short.width < 1080)
        assertTrue(short.height < 1920)
        assertEquals(2f / 3f, short.width.toFloat() / 1080f, 0.001f)

        val narrow = SplashLaunchGeometry.framePx(720, 1600, 2f)
        val fixed250 = (250f * 2f).roundToInt()
        assertEquals(480, narrow.width)
        assertTrue(narrow.width != fixed250)
        assertEquals((720f / 2f * (2f / 3f) * 2f).roundToInt(), narrow.width)

        val shrunk = sloganDrawPlacement(
            720f,
            1280f,
            measuredTextWidth = { fontSize -> fontSize * 20f },
            inkInsetPx = { 0f },
        )
        val unfitted = SplashLayout.sloganInFrame(720f, 1280f)
        assertTrue(shrunk.fontSizePx < unfitted.fontSize + 0.001f)
        assertEquals(
            fittedSloganFontSize(unfitted.fontSize, unfitted.fontSize * 20f, unfitted.width),
            shrunk.fontSizePx,
            0.001f,
        )

        val artwork = moduleFile("src/main/java/dev/qlutch/app/splash/SplashLaunchArtwork.kt").readText()
        assertTrue(artwork.contains("SplashLaunchGeometry.framePx"))
        assertTrue(artwork.contains("SplashLaunchGeometry.wordmarkInFramePx"))
        assertTrue(artwork.contains("sloganDrawPlacement"))
        assertTrue(artwork.contains("R.drawable.splash_first_frame"))
        assertTrue(artwork.contains("R.drawable.splash_g4_wordmark"))
        assertTrue(artwork.contains("R.string.splash_slogan"))
        assertTrue(artwork.contains("R.color.splash_brand"))
        assertTrue(artwork.contains("R.color.splash_background"))
        assertTrue(artwork.contains("containScale"))
        assertFalse(artwork.contains("R.raw.splash"))
        assertFalse(artwork.contains("250dp"))
        assertFalse(artwork.contains("444.444"))
        assertFalse(artwork.contains("变得轻松有趣"))

        val missingStill = SplashLaunchHandoff()
        missingStill.onStillLaidOut(0, 1280)
        missingStill.onWordmarkLaidOut(396, 100)
        missingStill.onSloganLaidOut(396, 40)
        assertFalse("a zero-size still is not on screen", missingStill.contentReady)

        val missingSlogan = SplashLaunchHandoff()
        missingSlogan.onStillLaidOut(720, 1280)
        missingSlogan.onWordmarkLaidOut(396, 100)
        assertFalse("the still without the slogan is not the in-app splash", missingSlogan.contentReady)

        val handoff = SplashLaunchHandoff()
        handoff.onStillLaidOut(720, 1280)
        handoff.onWordmarkLaidOut(0, 0)
        handoff.onSloganLaidOut(0, 0)
        assertFalse(handoff.contentReady)
        handoff.onWordmarkLaidOut(396, 100)
        assertFalse(handoff.contentReady)
        handoff.onSloganLaidOut(396, 40)
        assertTrue(handoff.contentReady)
        handoff.onStillLaidOut(0, 0)
        assertTrue("a later empty pass does not bring the system splash back", handoff.contentReady)

        val blocked = SplashDrawGate { false }
        assertFalse(blocked.onPreDraw())
        assertFalse(blocked.released)
        assertFalse(blocked.onPreDraw())
        var ready = false
        val gate = SplashDrawGate { ready }
        assertFalse(gate.onPreDraw())
        ready = true
        assertTrue(gate.onPreDraw())
        assertTrue(gate.released)
        ready = false
        assertTrue(gate.onPreDraw())

        val activity = moduleFile("src/main/java/dev/qlutch/app/splash/SplashActivity.kt").readText()
        assertTrue(
            activity.indexOf("holdSplashUntilContentReady") < activity.indexOf("setContent {"),
        )
        assertTrue(activity.indexOf("SplashLaunchArtworkDrawable") < activity.indexOf("setContent {"))
        assertTrue(activity.indexOf("setOnExitAnimationListener") < activity.indexOf("setContent {"))
        assertTrue(activity.contains("splashScreenView.remove()"))
        assertFalse(activity.contains("ObjectAnimator"))
        assertFalse(activity.contains("icon_preferred"))
        assertTrue(activity.contains("launchHandoff.contentReady"))
        assertTrue(activity.contains("LocalSplashLaunchHandoff provides launchHandoff"))
        assertTrue(activity.contains("StaticWelcomeStill(onContinue = ::openNext)"))
        assertTrue(activity.contains("ColdStartVideo(onFinished = ::openNext)"))
        val video = moduleFile("src/main/java/dev/qlutch/app/splash/SplashVideoView.kt").readText()
        val still = moduleFile("src/main/java/dev/qlutch/app/splash/StaticWelcomeStill.kt").readText()
        val frame = moduleFile("src/main/java/dev/qlutch/app/splash/SplashFrame.kt").readText()
        assertTrue(video.contains("handoff.onStillLaidOut(frameWidth, frameHeight)"))
        assertTrue(video.contains("handoff.onWordmarkLaidOut(markWidth, markHeight)"))
        assertTrue(video.contains("handoff.onSloganLaidOut(slogan.width, slogan.height)"))
        assertTrue(video.indexOf("layoutSplashSlogan(slogan, frameWidth, frameHeight)") < video.indexOf("handoff.onSloganLaidOut"))
        assertTrue(still.contains("launchHandoff.onStillLaidOut(size.width, size.height)"))
        assertTrue(still.contains("launchHandoff.onSloganLaidOut(width, height)"))
        assertTrue(still.contains("delay(SplashTiming.STATIC_HOLD_MILLIS)"))
        assertFalse(still.contains("POST_COMPLETION_HOLD_MILLIS"))
        assertTrue(frame.contains("launchHandoff.onWordmarkLaidOut(size.width, size.height)"))
        assertEquals(2f / 3f, SplashLayout.WIDTH_FRACTION, 0f)
        assertEquals(1_600L, SplashTiming.STATIC_HOLD_MILLIS)
        assertEquals(1_000L, SplashTiming.POST_COMPLETION_HOLD_MILLIS)
    }

    private class HoldClock {
        private var now = 0L
        val delays = mutableListOf<Long>()
        private val tasks = mutableListOf<Scheduled>()

        private class Scheduled(
            val due: Long,
            val action: () -> Unit,
            var cancelled: Boolean = false,
        )

        fun schedule(delayMillis: Long, action: () -> Unit): SplashCompletionHold.Cancellable {
            delays += delayMillis
            val scheduled = Scheduled(now + delayMillis, action)
            tasks += scheduled
            return SplashCompletionHold.Cancellable { scheduled.cancelled = true }
        }

        fun advance(millis: Long) {
            check(millis >= 0L)
            now += millis
            val due = tasks.filter { !it.cancelled && it.due <= now }
            tasks.removeAll(due)
            due.forEach { it.action() }
        }
    }

    private fun launcherHashes(hashes: Map<String, String>, dir: String): String = hashes.getValue(dir)

    private fun assertEquals(expected: Int, actual: Int, tolerance: Int) {
        assertTrue("expected $expected ±$tolerance but was $actual", kotlin.math.abs(expected - actual) <= tolerance)
    }

    private fun ByteArray.containsSequence(needle: ByteArray): Boolean {
        if (needle.isEmpty() || needle.size > size) return false
        for (start in 0..size - needle.size) {
            var matched = true
            for (offset in needle.indices) {
                if (this[start + offset] != needle[offset]) {
                    matched = false
                    break
                }
            }
            if (matched) return true
        }
        return false
    }

    @Test
    fun launcherIsThePortraitSplashAndTheShellStaysNext() {
        val manifest = moduleFile("src/main/AndroidManifest.xml").readText()
        assertTrue(manifest.contains("android:name=\".splash.SplashActivity\""))
        assertTrue(manifest.contains("android:screenOrientation=\"portrait\""))
        assertTrue(manifest.contains("android.intent.action.MAIN"))
        assertTrue(manifest.contains("android:name=\".MainActivity\""))
        val shell = moduleFile("src/main/java/dev/qlutch/app/MainActivity.kt").readText()
        assertTrue(shell.contains("首页"))
        assertTrue(shell.contains("做题"))
        assertTrue(shell.contains("PK"))
        assertTrue(shell.contains("好友"))
        assertTrue(shell.contains("我的"))
    }

    private fun sha256(bytes: ByteArray): String {
        val digest = MessageDigest.getInstance("SHA-256").digest(bytes)
        return digest.joinToString("") { "%02x".format(it) }
    }

    private class PngImage(val width: Int, val height: Int, private val pixels: IntArray) {
        fun alpha(x: Int, y: Int) = (pixels[y * width + x] ushr 24) and 0xff
        fun red(x: Int, y: Int) = (pixels[y * width + x] ushr 16) and 0xff
        fun green(x: Int, y: Int) = (pixels[y * width + x] ushr 8) and 0xff
        fun blue(x: Int, y: Int) = pixels[y * width + x] and 0xff
    }

    private fun decodePng(bytes: ByteArray): PngImage {
        val signature = byteArrayOf(
            0x89.toByte(), 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
        )
        require(bytes.copyOfRange(0, 8).contentEquals(signature))
        var pos = 8
        var width = 0
        var height = 0
        var colorType = -1
        val idat = java.io.ByteArrayOutputStream()
        while (pos + 12 <= bytes.size) {
            val length = readInt(bytes, pos)
            val type = bytes.copyOfRange(pos + 4, pos + 8).toString(Charsets.US_ASCII)
            val start = pos + 8
            when (type) {
                "IHDR" -> {
                    width = readInt(bytes, start)
                    height = readInt(bytes, start + 4)
                    require(bytes[start + 8].toInt() and 0xff == 8)
                    colorType = bytes[start + 9].toInt() and 0xff
                    require(bytes[start + 12].toInt() and 0xff == 0)
                }
                "IDAT" -> idat.write(bytes, start, length)
                "IEND" -> break
            }
            pos = start + length + 4
        }
        val channels = when (colorType) {
            6 -> 4
            2 -> 3
            else -> error("unsupported color type $colorType")
        }
        val stride = width * channels
        val expected = height * (1 + stride)
        val raw = ByteArray(expected)
        val inflater = java.util.zip.Inflater()
        inflater.setInput(idat.toByteArray())
        var produced = 0
        while (produced < expected) {
            val count = inflater.inflate(raw, produced, expected - produced)
            check(count > 0)
            produced += count
        }
        inflater.end()
        val pixels = IntArray(width * height)
        val prev = IntArray(stride)
        val row = IntArray(stride)
        var cursor = 0
        for (y in 0 until height) {
            val filter = raw[cursor].toInt() and 0xff
            cursor += 1
            for (x in 0 until stride) {
                row[x] = raw[cursor].toInt() and 0xff
                cursor += 1
            }
            when (filter) {
                0 -> Unit
                1 -> for (x in 0 until stride) {
                    val left = if (x >= channels) row[x - channels] else 0
                    row[x] = (row[x] + left) and 0xff
                }
                2 -> for (x in 0 until stride) {
                    row[x] = (row[x] + prev[x]) and 0xff
                }
                3 -> for (x in 0 until stride) {
                    val left = if (x >= channels) row[x - channels] else 0
                    row[x] = (row[x] + ((left + prev[x]) / 2)) and 0xff
                }
                4 -> for (x in 0 until stride) {
                    val left = if (x >= channels) row[x - channels] else 0
                    val up = prev[x]
                    val upLeft = if (x >= channels) prev[x - channels] else 0
                    val estimate = left + up - upLeft
                    val leftDelta = kotlin.math.abs(estimate - left)
                    val upDelta = kotlin.math.abs(estimate - up)
                    val upLeftDelta = kotlin.math.abs(estimate - upLeft)
                    val predictor = when {
                        leftDelta <= upDelta && leftDelta <= upLeftDelta -> left
                        upDelta <= upLeftDelta -> up
                        else -> upLeft
                    }
                    row[x] = (row[x] + predictor) and 0xff
                }
                else -> error("png filter $filter")
            }
            for (x in 0 until width) {
                val offset = x * channels
                val red = row[offset]
                val green = row[offset + 1]
                val blue = row[offset + 2]
                val alpha = if (channels == 4) row[offset + 3] else 255
                pixels[y * width + x] = (alpha shl 24) or (red shl 16) or (green shl 8) or blue
            }
            row.copyInto(prev)
        }
        return PngImage(width, height, pixels)
    }

    private fun readInt(bytes: ByteArray, offset: Int): Int {
        return ((bytes[offset].toInt() and 0xff) shl 24) or
            ((bytes[offset + 1].toInt() and 0xff) shl 16) or
            ((bytes[offset + 2].toInt() and 0xff) shl 8) or
            (bytes[offset + 3].toInt() and 0xff)
    }

    private fun moduleFile(relative: String): File {
        return moduleFileIfExists(relative)
            ?: error("Missing $relative from ${File(System.getProperty("user.dir") ?: ".").absolutePath}")
    }

    private fun moduleFileIfExists(relative: String): File? {
        val cwd = File(System.getProperty("user.dir") ?: ".")
        val candidates = listOf(
            File(cwd, relative),
            File(cwd, "app/$relative"),
            File(cwd, "apps/android/app/$relative"),
        )
        return candidates.firstOrNull { it.exists() }
    }
}

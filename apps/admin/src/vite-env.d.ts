interface ImportMetaEnv {
  readonly VITE_QLUTCH_ENV?: string
  readonly VITE_QLUTCH_API_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

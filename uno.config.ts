import { readFile } from 'node:fs/promises'
import { defineConfig, presetWind4, transformerVariantGroup } from 'unocss'

const path = "./node_modules/@unocss/reset/tailwind-compat.css"


export default defineConfig({
    cli: {
        entry: {
            patterns: [
                "./web/**/*.templ",
            ],
            outFile: "./web/static/index.css"
        }
    },
    preflights: [
        {
            layer: "base",
            getCSS: () => readFile(path, "utf-8")
        }
    ],
    presets: [
        presetWind4()
    ],
    transformers: [
        transformerVariantGroup()
    ]
})

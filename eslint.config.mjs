import antfu from '@antfu/eslint-config'

// Preserve the upstream coding style while using ESLint's supported flat config.
export default antfu({
  vue: true,
  stylistic: false,
  markdown: false,
  yaml: false,
  jsonc: false,
  unicorn: false,
  ignores: ['service/**', 'dist/**', 'node_modules/**'],
  rules: {
    'no-console': 'off',
    'vue/html-indent': 'off',
    'vue/space-infix-ops': 'off',
    'vue/block-order': 'off',
    'e18e/prefer-date-now': 'off',
    'perfectionist/sort-imports': 'off',
    'perfectionist/sort-named-imports': 'off',
    'perfectionist/sort-named-exports': 'off',
    'node/prefer-global/process': 'off',
    'perfectionist/sort-exports': 'off',
    'ts/method-signature-style': 'off',
    'ts/no-duplicate-enum-values': 'off', // Upstream enum keeps aliases.
    'import/consistent-type-specifier-style': 'off',
    'unused-imports/no-unused-vars': ['warn', { args: 'none', varsIgnorePattern: '^_' }],
  },
})

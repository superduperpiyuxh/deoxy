import * as esbuild from 'esbuild';

const production = process.argv.includes('--production');
const watch = process.argv.includes('--watch');

async function main() {
  const ctx = await esbuild.context({
    entryPoints: ['src/extension.ts'],
    outfile: 'dist/extension.js',
    bundle: true,
    platform: 'node',
    format: 'cjs',
    external: ['vscode'],
    sourcemap: production ? false : 'inline',
    minify: production,
    target: 'es2020',
  });

  if (watch) {
    await ctx.watch();
    console.log('[deoxy] watching for changes...');
  } else {
    await ctx.rebuild();
    console.log('[deoxy] build complete');
    await ctx.dispose();
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});

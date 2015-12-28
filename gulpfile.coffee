gulp = require 'gulp'
rename = require 'gulp-rename'

gulp.task 'copy', ->
  gulp.src [
      'bower_components/jquery/dist/jquery.min.js',
      'bower_components/jquery/dist/jquery.min.map',
      'bower_components/emojify.js/dist/js/emojify.min.js',
      'bower_components/toastr/toastr.min.js',
      'bower_components/highlightjs/highlight.pack.min.js'
    ]
    .pipe gulp.dest 'build/public/js'
  gulp.src [
      'bower_components/toastr/toastr.min.css',
      'bower_components/highlightjs/styles/github.css'
    ]
    .pipe gulp.dest 'build/public/css'
  gulp.src [
      'bower_components/emojione/emoji.json'
      'bower_components/emojione/emoji_strategy.json'
      'bower_components/emojione/lib/js/emojione.min.js'
    ]
    .pipe gulp.dest 'build/public/js'
  gulp.src 'bower_components/emojione/assets/css/emojione.min.css'
    .pipe gulp.dest 'build/public/css'
  gulp.src 'bower_components/emojione/assets/png/*'
    .pipe gulp.dest 'build/public/images/emojione'
  gulp.src [
      'public/js/**',
      'public/css/*',
      'public/fonts/*',
      'public/ico/*'
    ], base: 'public'
    .pipe gulp.dest 'build/public'
  gulp.src 'template/*'
    .pipe gulp.dest 'build/template'

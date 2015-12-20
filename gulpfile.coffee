gulp = require 'gulp'

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
  gulp.src 'bower_components/emojify.js/dist/css/sprites/emojify.min.css'
    .pipe gulp.dest 'build/public/css/emojify.sprites.min.css'
  gulp.src 'bower_components/emojify.js/dist/css/basic/emojify.min.css'
    .pipe gulp.dest 'build/public/css/emojify.basic.min.css'
  gulp.src 'bower_components/emojify.js/dist/images/sprites/*'
    .pipe gulp.dest 'build/public/images'
  gulp.src 'bower_components/emojify.js/dist/images/basic/*'
    .pipe gulp.dest 'build/public/images/emoji'
  gulp.src [
      'public/js/**',
      'public/css/*',
      'public/fonts/*',
      'public/ico/*'
    ], { base: 'public' }
    .pipe gulp.dest 'build/public'
  gulp.src 'template/*'
    .pipe gulp.dest 'build/template'

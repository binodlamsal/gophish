/*
 * gulpfile.js
 *
 * Description: The Gophish gulpfile
 */

var gulp = require("gulp"),
  rename = require("gulp-rename"),
  concat = require("gulp-concat"),
  uglify = require("gulp-uglify"),
  cleanCSS = require("gulp-clean-css"),
  js_directory = "static/js/src/",
  css_directory = "static/css/",
  vendor_directory = js_directory + "vendor/",
  app_directory = js_directory + "app/**/*.js",
  dest_js_directory = "static/js/dist/",
  dest_css_directory = "static/css/dist/";

gulp.task("vendorjs", function() {
  // Vendor minifying / concat
  return gulp
    .src([
      vendor_directory + "jquery.js",
      vendor_directory + "bootstrap.min.js",
      vendor_directory + "moment.min.js",
      vendor_directory + "papaparse.min.js",
      vendor_directory + "d3.min.js",
      vendor_directory + "topojson.min.js",
      vendor_directory + "datamaps.min.js",
      vendor_directory + "jquery.dataTables.min.js",
      vendor_directory + "dataTables.bootstrap.js",
      vendor_directory + "dataTables.select.min.js",
      vendor_directory + "datetime-moment.js",
      vendor_directory + "jquery.ui.widget.js",
      vendor_directory + "jquery.fileupload.js",
      vendor_directory + "jquery.iframe-transport.js",
      vendor_directory + "sweetalert2.min.js",
      vendor_directory + "bootstrap-datetime.js",
      vendor_directory + "select2.min.js",
      vendor_directory + "core.min.js",
      vendor_directory + "highcharts.js",
      vendor_directory + "ua-parser.min.js"
    ])
    .pipe(concat("vendor.js"))
    .pipe(
      rename({
        suffix: ".min"
      })
    )
    .pipe(uglify())
    .pipe(gulp.dest(dest_js_directory));
});

gulp.task("scripts", function() {
  // Gophish app files
  gulp
    .src(app_directory)
    .pipe(
      rename({
        suffix: ".min"
      })
    )
    .pipe(
      uglify().on("error", function(e) {
        console.log(e);
      })
    )
    .pipe(gulp.dest(dest_js_directory + "app/"));
});

gulp.task("styles", function() {
  return gulp
    .src([
      css_directory + "bootstrap.min.css",
      css_directory + "main.css",
      css_directory + "dashboard.css",
      css_directory + "flat-ui.css",
      css_directory + "dataTables.bootstrap.css",
      css_directory + "select.dataTables.min.css",
      css_directory + "font-awesome.min.css",
      css_directory + "chartist.min.css",
      css_directory + "bootstrap-datetime.css",
      css_directory + "checkbox.css",
      css_directory + "sweetalert2.min.css",
      css_directory + "select2.min.css",
      css_directory + "select2-bootstrap.min.css"
    ])
    .pipe(
      cleanCSS({
        compatibilty: "ie9"
      })
    )
    .pipe(concat("gophish.css"))
    .pipe(gulp.dest(dest_css_directory));
});

gulp.task("build", ["vendorjs", "scripts", "styles"]);

gulp.task("default", ["build"]);

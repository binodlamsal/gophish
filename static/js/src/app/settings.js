$(document).ready(function() {
  $("#apiResetForm").submit(function(e) {
    return (
      api
        .reset()
        .success(function(e) {
          (user.api_key = e.data),
            successFlash(e.message),
            $("#api_key").val(user.api_key);
        })
        .error(function(e) {
          errorFlash(e.message);
        }),
      !1
    );
  });

  $("#settingsForm").submit(function(e) {
    e.preventDefault();

    return (
      $.post("/settings", $(this).serialize())
        .done(function(e) {
          successFlash(e.message);
        })
        .fail(function(e) {
          errorFlash(e.responseJSON.message);
        }),
      !1
    );
  });

  var e = localStorage.getItem("gophish.use_map");

  $("#use_map").prop("checked", JSON.parse(e)),
    $("#use_map").on("change", function() {
      localStorage.setItem("gophish.use_map", JSON.stringify(this.checked));
    });

  $(document).on("change", ".btn-file .logo :file", function() {
    var input = $(this),
      label = input
        .val()
        .replace(/\\/g, "/")
        .replace(/.*\//, "");
    input.trigger("fileselect", [label]);
  });

  $(document).on("change", ".btn-file .avatar :file", function() {
    var input = $(this),
      label = input
        .val()
        .replace(/\\/g, "/")
        .replace(/.*\//, "");
    input.trigger("fileselect", [label]);
  });

  $(".btn-file .logo :file").on("fileselect", function(event, label) {
    var input = $(this)
        .parents(".input-group")
        .find(":text"),
      log = label;

    if (input.length) {
      input.val(log);
    } else {
      if (log) alert(log);
    }
  });

  $(".btn-file .avatar :file").on("fileselect", function(event, label) {
    var input = $(this)
        .parents(".input-group")
        .find(":text"),
      log = label;

    if (input.length) {
      input.val(log);
    } else {
      if (log) alert(log);
    }
  });

  function readURL(input, type) {
    if (input.files && input.files[0]) {
      if (input.files[0].size * 1.4 > 65000) {
        alert(
          "The " + type + " image file is too large (must not exceed 64KB)"
        );
        return;
      }

      var reader = new FileReader();

      reader.onload = function(e) {
        $("#" + type + "-preview").attr("src", e.target.result);
        $("#" + type).val(e.target.result);
      };

      reader.readAsDataURL(input.files[0]);
    }
  }

  $("#logo-input").change(function() {
    readURL(this, "logo");
  });

  $("#avatar-input").change(function() {
    readURL(this, "avatar");
  });

  $("button#reset-logo").click(function(e) {
    e.preventDefault();

    $("#logo-preview").attr(
      "src",
      "data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
    );
    $("#logo").val("DELETE");
  });

  $("button#reset-avatar").click(function(e) {
    e.preventDefault();

    $("#avatar-preview").attr(
      "src",
      "data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
    );
    $("#avatar").val("DELETE");
  });
});

var pagesTable;

function save(e) {
  var a = {};
  (a.name = $("#name").val()),
    (a.tag = parseInt($("#category").val())),
    (a.public = $("#publicly_available").prop("checked")),
    (editor = CKEDITOR.instances.html_editor),
    (a.html = editor.getData()),
    (a.capture_credentials = $("#capture_credentials_checkbox").prop(
      "checked"
    )),
    (a.capture_passwords = $("#capture_passwords_checkbox").prop("checked")),
    (a.redirect_url = $("#redirect_url_input").val()),
    -1 != e
      ? ((a.id = pages[e].id),
        api.pageId.put(a).success(function(e) {
          successFlash("Page edited successfully!"),
            load($("input[type=radio][name=filter]:checked").val()),
            dismiss();
        }))
      : api.pages
          .post(a)
          .success(function(e) {
            successFlash("Page added successfully!"),
              load($("input[type=radio][name=filter]:checked").val()),
              dismiss();
          })
          .error(function(e) {
            modalError(e.responseJSON.message);
          });
}

function dismiss() {
  $("#modal\\.flashes").empty(),
    $("#name").val(""),
    $("#html_editor").val(""),
    $("#url").val(""),
    $("#redirect_url_input").val(""),
    $("#modal")
      .find("input[type='checkbox']")
      .prop("checked", !1),
    $("#capture_passwords").hide(),
    $("#redirect_url").hide(),
    $("#modal").modal("hide");
}

function importSite() {
  (url = $("#url").val()),
    url
      ? api
          .clone_site({
            url: url,
            include_resources: !1
          })
          .success(function(e) {
            $("#html_editor").val(e.html),
              CKEDITOR.instances.html_editor.setMode("wysiwyg"),
              $("#importSiteModal").modal("hide");
          })
          .error(function(e) {
            modalError(e.responseJSON.message);
          })
      : modalError("No URL Specified!");
}

function edit(e) {
  $("#modalSubmit")
    .unbind("click")
    .click(function() {
      save(e);
    }),
    $("#html_editor").ckeditor();
  var a = {};
  -1 != e &&
    ((a = pages[e]),
    $("#name").val(a.name),
    $("#html_editor").val(a.html),
    $("#publicly_available").prop("checked", a.public),
    $("#capture_credentials_checkbox").prop("checked", a.capture_credentials),
    $("#capture_passwords_checkbox").prop("checked", a.capture_passwords),
    $("#redirect_url_input").val(a.redirect_url),
    a.capture_credentials &&
      ($("#capture_passwords").show(), $("#redirect_url").show()));

  //fill the categories by the API
  $("#category")
    .find("option")
    .not(":first")
    .remove();
  api.phishtags.get().success(function(s) {
    $.each(s, function(e, ss) {
      var sel = "";
      if (a.tag == ss.id) {
        sel = 'selected = "selected"';
      }

      $("#category").append(
        '<option value="' + ss.id + '"  ' + sel + ">" + ss.name + "</option>"
      );
    });
  });
}

function copy(e) {
  $("#modalSubmit")
    .unbind("click")
    .click(function() {
      save(-1);
    }),
    $("#html_editor").ckeditor();
  var a = pages[e];
  $("#name").val("Copy of " + a.name), $("#html_editor").val(a.html);
}

function load(filter) {
  if (pagesTable === undefined) {
    pagesTable = $("#pagesTable").DataTable({
      destroy: !0,
      columnDefs: [
        {
          orderable: !1,
          targets: "no-sort"
        }
      ]
    });
    $("#pagesTable").show();
  } else {
    pagesTable.clear();
    pagesTable.draw();
  }

  $("#loading").show(),
    api.pages
      .get(filter)
      .success(function(e) {
        (pages = e),
          $("#loading").hide(),
          pages.length > 0
            ? ($("#pagesTable").show(),
              $.each(pages, function(e, a) {
                pagesTable.row
                  .add([
                    escapeHtml(a.name),
                    a.username,
                    moment(a.modified_date).format("MMMM Do YYYY, h:mm:ss a"),
                    "<div class='pull-right'><span data-toggle='modal' data-backdrop='static' data-target='#modal'>" +
                      (a.writable
                        ? "<button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Edit Page' onclick='edit(" +
                          e +
                          ")'><i class='fa fa-pencil'></i></button></span>\t\t"
                        : "") +
                      "  <span data-toggle='modal' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Copy Page' onclick='copy(" +
                      e +
                      ")'><i class='fa fa-copy'></i></button></span>\t\t" +
                      (a.writable
                        ? "<button class='btn btn-danger' data-toggle='tooltip' data-placement='left' title='Delete Page' onclick='deletePage(" +
                          e +
                          ")'><i class='fa fa-trash-o'></i></button>"
                        : "") +
                      "</div>"
                  ])
                  .draw();
              }),
              $('[data-toggle="tooltip"]').tooltip())
            : $("#emptyMessage").hide();
      })
      .error(function() {
        $("#loading").hide(), errorFlash("Error fetching pages");
      });
}
var pages = [],
  deletePage = function(e) {
    swal({
      title: "Are you sure?",
      text: "This will delete the landing page. This can't be undone!",
      type: "warning",
      animation: !1,
      showCancelButton: !0,
      confirmButtonText: "Delete " + escapeHtml(pages[e].name),
      confirmButtonColor: "#428bca",
      reverseButtons: !0,
      allowOutsideClick: !1,
      preConfirm: function() {
        return new Promise(function(a, t) {
          api.pageId
            .delete(pages[e].id)
            .success(function(e) {
              a();
            })
            .error(function(e) {
              t(e.responseJSON.message);
            });
        });
      }
    }).then(function() {
      swal(
        "Landing Page Deleted!",
        "This landing page has been deleted!",
        "success"
      ),
        $('button:contains("OK")').on("click", function() {
          location.reload();
        });
    });
  };
$(document).ready(function() {
  $(".modal").on("hidden.bs.modal", function(e) {
    $(this).removeClass("fv-modal-stack"),
      $("body").data("fv_open_modals", $("body").data("fv_open_modals") - 1);
  }),
    $(".modal").on("shown.bs.modal", function(e) {
      void 0 === $("body").data("fv_open_modals") &&
        $("body").data("fv_open_modals", 0),
        $(this).hasClass("fv-modal-stack") ||
          ($(this).addClass("fv-modal-stack"),
          $("body").data(
            "fv_open_modals",
            $("body").data("fv_open_modals") + 1
          ),
          $(this).css("z-index", 1040 + 10 * $("body").data("fv_open_modals")),
          $(".modal-backdrop")
            .not(".fv-modal-stack")
            .css("z-index", 1039 + 10 * $("body").data("fv_open_modals")),
          $(".modal-backdrop")
            .not("fv-modal-stack")
            .addClass("fv-modal-stack"));
    }),
    ($.fn.modal.Constructor.prototype.enforceFocus = function() {
      $(document)
        .off("focusin.bs.modal")
        .on(
          "focusin.bs.modal",
          $.proxy(function(e) {
            this.$element[0] === e.target ||
              this.$element.has(e.target).length ||
              $(e.target).closest(".cke_dialog, .cke").length ||
              this.$element.trigger("focus");
          }, this)
        );
    }),
    $(document).on("hidden.bs.modal", ".modal", function() {
      $(".modal:visible").length && $(document.body).addClass("modal-open");
    }),
    $("#modal").on("hidden.bs.modal", function(e) {
      dismiss();
    }),
    $("#capture_credentials_checkbox").change(function() {
      $("#capture_passwords").toggle(), $("#redirect_url").toggle();
    }),
    $("input[type=radio][name=filter]").change(function(event) {
      load(event.target.value);
    });

  load("own");
});

// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func WavesBackground() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<style>\n  body {\n    /* background-color: #5461C3 */\n    background-image: linear-gradient(to right, #3678f9, #5461C3)\n  }\n\n  .box {\n    position: fixed;\n    top: 0;\n    transform: rotate(40deg);\n    left: 0;\n  }\n\n  .wave {\n    position: absolute;\n    opacity: .5;\n    width: 1500px;\n    height: 1300px;\n    margin-left: -150px;\n    margin-top: -250px;\n    border-radius: 43%;\n  }\n\n  @keyframes rotate {\n    from {\n      transform: rotate(0deg);\n    }\n\n    from {\n      transform: rotate(360deg);\n    }\n  }\n\n  .wave.-one {\n    animation: rotate 10000ms infinite linear;\n    background: #1c766c;\n  }\n\n  .wave.-two {\n    animation: rotate 6000ms infinite linear;\n    background: #5dffff;\n  }\n\n  .star {\n    background-image: url('./kuz.png');\n  }\n</style><div class=\"box\"><div class=\"wave -one\"></div><div class=\"wave -two\"></div><!-- <div class='star'></div> --></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate

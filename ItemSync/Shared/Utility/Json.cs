using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http.Formatting;
using System.Text;
using System.Threading.Tasks;
using System.Web.Http;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;

namespace ItemSync.Shared.Utility
{

    public static class Json {
        public static JsonSerializerSettings JsonSerializerSettings => new JsonSerializerSettings {
            ContractResolver = new CamelCasePropertyNamesContractResolver(),
            Formatting = Formatting.Indented,
            NullValueHandling = NullValueHandling.Ignore
        };

        public static MediaTypeFormatter JsonFormatter = new JsonMediaTypeFormatter {
            SerializerSettings = JsonSerializerSettings
        };
    }
}

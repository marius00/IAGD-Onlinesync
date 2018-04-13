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

        public static List<MediaTypeFormatter> JsonFormatter = new List<MediaTypeFormatter> {
            new JsonMediaTypeFormatter {
                SerializerSettings = JsonSerializerSettings
            }
        };



        public static HttpConfiguration Config {
            get {
                var config = new HttpConfiguration();
                config.Formatters.JsonFormatter.SerializerSettings.ContractResolver = new CamelCasePropertyNamesContractResolver();
                config.Formatters.JsonFormatter.UseDataContractJsonSerializer = false;
                return config;
            }
        }
    }
}

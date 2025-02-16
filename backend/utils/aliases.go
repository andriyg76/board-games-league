package utils

import (
	"github.com/andriyg76/glog"
	"math/rand"
	"strings"
)

var subjects = []string{
	"swift", "brave", "mighty", "clever", "wise", "silent", "rapid", "fierce", "gentle", "bold",
	"active", "agile", "alert", "ancient", "animated", "ardent", "astute", "athletic", "august", "autumn",
	"balanced", "beaming", "blazing", "blessed", "blissful", "blithe", "blooming", "bonny", "boundless", "bright",
	"brilliant", "buoyant", "calm", "capable", "caring", "casual", "certain", "chief", "choice", "classic",
	"clear", "cosmic", "crisp", "crystal", "daring", "dauntless", "dawn", "dazzling", "dear", "decent",
	"deep", "definite", "delicate", "determined", "diligent", "divine", "dynamic", "eager", "early", "earnest",
	"easy", "electric", "elegant", "elite", "endless", "energetic", "eternal", "ethical", "exact", "exotic",
	"fair", "faithful", "famous", "fancy", "fast", "fine", "firm", "first", "fit", "fleet",
	"flowing", "flying", "fond", "frank", "free", "fresh", "friendly", "full", "funny", "gallant",
	"gathered", "giving", "glad", "glorious", "glowing", "golden", "good", "graceful", "grand", "grateful",
	"great", "green", "growing", "happy", "hardy", "harmonic", "healthy", "helpful", "hidden", "high",
	"honest", "hopeful", "humble", "ideal", "infinite", "innocent", "inspired", "integral", "intense", "intimate",
	"intuitive", "inviting", "jaunty", "jolly", "joyful", "joyous", "jubilant", "judicious", "keen", "kind",
	"knowing", "learned", "legal", "light", "lighting", "likely", "lively", "logical", "loving", "loyal",
	"lucky", "lunar", "magic", "magnetic", "main", "major", "merry", "mighty", "mindful", "natural",
	"neat", "noble", "normal", "noted", "novel", "organic", "original", "pacific", "patient", "peaceful",
	"perfect", "phoenix", "plain", "pleasant", "pleased", "plucky", "poetic", "pointed", "polite", "positive",
	"possible", "powerful", "precious", "precise", "prepared", "present", "pretty", "prime", "proper", "proud",
	"pure", "quick", "quiet", "radical", "rapid", "rare", "ready", "real", "refined", "regular",
	"relaxed", "reliable", "resolved", "rich", "right", "robust", "rooted", "royal", "sacred", "safe",
	"sage", "sane", "scenic", "secret", "secure", "select", "settled", "sharing", "sharp", "shining",
	"simple", "sincere", "single", "skilled", "smooth", "social", "solid", "sound", "special", "spiral",
	"spirit", "spring", "square", "stable", "steady", "stellar", "strict", "strong", "studious", "style",
	"subtle", "summer", "sunny", "super", "sure", "sweet", "swift", "talented", "tender", "thankful",
	"thorough", "tidy", "timely", "tireless", "tough", "trained", "true", "trusted", "truthful", "useful",
	"valid", "valued", "vast", "verbal", "verdant", "verified", "viable", "vital", "vivid", "warm",
	"wealthy", "welcome", "well", "whole", "willing", "winter", "wise", "witty", "wonderful", "worthy",
	"zealous", "zesty",
}

var names = []string{
	"hawk", "wolf", "eagle", "bear", "fox", "owl", "tiger", "lion", "deer", "raven",
	"ace", "aero", "alpha", "angel", "apex", "apollo", "archer", "aries", "arrow", "atlas",
	"aurora", "avalon", "badger", "bandit", "baron", "beacon", "beast", "blade", "blaze", "bloom",
	"breeze", "bridge", "brook", "buck", "bullet", "burst", "buzz", "byte", "calf", "captain",
	"cascade", "castle", "cave", "chain", "chance", "chaos", "charm", "chase", "child", "cipher",
	"circle", "cliff", "cloud", "coast", "comet", "corona", "cosmos", "crest", "crow", "crown",
	"crystal", "cube", "cyber", "dagger", "dawn", "delta", "demon", "depth", "desert", "diamond",
	"digit", "dragon", "dream", "drift", "drone", "drop", "dusk", "dust", "echo", "edge",
	"ember", "empire", "engine", "epoch", "equinox", "essence", "ether", "falcon", "fang", "feather",
	"fiber", "field", "fire", "flame", "flash", "fleet", "flight", "flow", "fluke", "flux",
	"force", "forest", "forge", "frost", "fury", "gale", "gamma", "gate", "ghost", "giant",
	"glacier", "glade", "glide", "globe", "glow", "goose", "grace", "grid", "griffin", "grove",
	"guard", "guide", "gulf", "halo", "harbor", "heart", "hedge", "helix", "hero", "horizon",
	"horn", "hunter", "hyper", "ice", "image", "index", "iris", "isle", "jade", "jazz",
	"jet", "jewel", "keep", "knight", "lake", "lance", "legend", "lens", "light", "lightning",
	"link", "lunar", "lynx", "magic", "magna", "manta", "marble", "marine", "mars", "mask",
	"matrix", "maze", "meteor", "mind", "mint", "mist", "moon", "moose", "mountain", "myth",
	"nebula", "nerve", "net", "nexus", "night", "ninja", "north", "nova", "oasis", "omega",
	"onyx", "orbit", "orchid", "orion", "palm", "path", "peak", "pearl", "phantom", "phoenix",
	"pilot", "pine", "planet", "plasma", "plaza", "plume", "point", "polar", "portal", "prism",
	"proxy", "pulse", "puma", "quail", "quantum", "quartz", "quest", "quick", "radius", "rain",
	"ram", "range", "raptor", "ray", "reef", "rhythm", "ridge", "rift", "river", "road",
	"rock", "root", "rose", "rover", "ruby", "rush", "sage", "sail", "scale", "scout",
	"shade", "shadow", "shark", "shield", "shore", "sigma", "signal", "silk", "silver", "sky",
	"slate", "smoke", "snake", "snow", "solar", "solid", "sonic", "spark", "sphinx", "spider",
	"spike", "spirit", "spring", "storm", "stream", "strike", "sun", "swan", "swift", "sword",
	"taurus", "tech", "temple", "terra", "theta", "thor", "thunder", "tide", "tiger", "titan",
	"tone", "torch", "tower", "track", "trail", "trend", "trinity", "tropic", "valley", "vapor",
	"vector", "venom", "vertex", "vessel", "victor", "view", "viking", "vine", "viper", "vision",
	"void", "volt", "vortex", "wake", "ward", "wave", "way", "whale", "wind", "wing",
	"winter", "wolf", "wonder", "wood", "world", "wyvern", "xenon", "yard", "year", "yeti",
	"yield", "zen", "zero", "zeta", "zinc", "zone",
}

func GetAlias() string {
	subject := subjects[rand.Intn(len(subjects))]
	name := names[rand.Intn(len(names))]

	// Capitalize first letters
	subject = strings.ToUpper(subject[:1]) + subject[1:]
	name = strings.ToUpper(name[:1]) + name[1:]

	if rand.Intn(2) == 0 {
		return subject + " " + name
	}
	return name + " " + subject
}

func GetUniqueAlias(isUnique func(string) (bool, error)) (string, error) {
	errors := 0
	for i := 0; i < 1000; i++ {
		subject := subjects[rand.Intn(len(subjects))]
		name := names[rand.Intn(len(names))]

		// Capitalize first letters
		subject = strings.ToUpper(subject[:1]) + subject[1:]
		name = strings.ToUpper(name[:1]) + name[1:]

		var alias string
		if rand.Intn(2) == 0 {
			alias = subject + " " + name
		} else {
			alias = name + " " + subject
		}
		if u, err := isUnique(alias); err != nil {
			errors = errors + 1
			_ = glog.Error("error happened while alias checked for uniqueness")
			continue
		} else if u {
			return alias, nil
		}
	}
	return "", glog.Error("can't generate unique alias, %d errors happened", errors)
}
